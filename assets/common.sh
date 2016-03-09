export TMPDIR=${TMPDIR:-/tmp}

load_pubkey() {
  local private_key_path=$TMPDIR/git-resource-private-key

  (jq -r '.source.private_key // empty' < $1) > $private_key_path

  if [ -s $private_key_path ]; then
    chmod 0600 $private_key_path

    eval $(ssh-agent) >/dev/null 2>&1
    trap "kill $SSH_AGENT_PID" 0

    SSH_ASKPASS=/opt/resource/askpass.sh DISPLAY= ssh-add $private_key_path >/dev/null

    mkdir -p ~/.ssh
    cat > ~/.ssh/config <<EOF
StrictHostKeyChecking no
LogLevel quiet
EOF
    chmod 0600 ~/.ssh/config
  fi
}

configure_git_ssl_verification() {
  skip_ssl_verification=$(jq -r '.source.skip_ssl_verification // false' < $1)
  if [ "$skip_ssl_verification" = "true" ]; then
    export GIT_SSL_NO_VERIFY=true
  fi
}

hg_metadata() {
  local ref=$(hg identify --id)

  local commit=$(hg log --rev $ref --template "{node}" | jq -R .)
  local author=$(hg log --rev $ref --template "{author}" | jq -s -R .)
  local author_date=$(hg log --rev $ref --template "{date|isodatesec}" | jq -R .)
  local message=$(hg log --rev $ref --template "{desc}" | jq -s -R .)

  jq -n "[
    {name: \"commit\", value: ${commit}},
    {name: \"author\", value: ${author}},
    {name: \"author_date\", value: ${author_date}, type: \"time\"},
    {name: \"message\", value: ${message}, type: \"message\"}
  ]"
}

check_revision_exists() {
  hg log --rev $1 &>/dev/null
}