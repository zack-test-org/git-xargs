#!/usr/bin/env bash 

echo "Upgrading CircleCI workflows syntax to 2..."

 yq w -i .circleci/config.yml 'workflows.version' 2

 # yq has an annoying habit of adding stray merge tags to the final output
 sed -i '/!!merge /d' .circleci/config.yml
