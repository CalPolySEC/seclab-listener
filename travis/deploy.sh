#!/bin/bash
set -e
eval "$(ssh-agent -s)"
chmod 700 travis
openssl aes-256-cbc -K $encrypted_fa51861454fb_key -iv $encrypted_fa51861454fb_iv -in travis/deploy.enc -out travis/deploy -d
chmod 600 travis/deploy
ssh-add travis/deploy
git remote add deploy "git@thewhitehat.club:go/src/github.com/WhiteHatCP/seclab-listener"
git push deploy
