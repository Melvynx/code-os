import{n as e,r as t,t as n}from"./skills-sync-docs-D0By4tkc.js";import{i as r}from"./index-CZop_SAX.js";var i=r(),a=[{title:`Back up the VPS library`,children:(0,i.jsxs)(i.Fragment,{children:[(0,i.jsx)(`p`,{children:`Treat the existing VPS library as the initial source of truth. Make a timestamped copy before adding Git metadata.`}),(0,i.jsx)(n,{children:`test -d ~/.agents
backup_dir="$HOME/.agents.backup.$(date +%Y%m%d-%H%M%S)"
cp -a ~/.agents "$backup_dir"
echo "Backup: $backup_dir"`}),(0,i.jsx)(t,{children:`Nothing is deleted or moved. Keep this backup until both machines have completed several successful syncs.`})]})},{title:`Create the private Git source`,children:(0,i.jsxs)(i.Fragment,{children:[(0,i.jsx)(`p`,{children:`Authenticate GitHub CLI over HTTPS first. Its credential helper remains available to the background timer without an SSH agent.`}),(0,i.jsx)(n,{children:`gh auth status || gh auth login
gh config set git_protocol https
gh auth setup-git`}),(0,i.jsxs)(`p`,{children:[`Then initialize `,(0,i.jsx)(`code`,{children:`~/.agents`}),` directly so the repository contains the library contents—not an extra wrapper directory. Exclude credentials before the first commit.`]}),(0,i.jsx)(n,{children:`cd ~/.agents
test -d .git || git init -b main
printf '%s
' '.env' '.env.*' '*.key' '*.pem' '*token*' '*secret*' >> .gitignore
git config user.name "YOUR_NAME"
git config user.email "YOUR_EMAIL"
git add -A
git commit -m "chore: initialize shared agent skills"
gh repo create code-os-skills --private --source=. --remote=origin --push`}),(0,i.jsx)(t,{security:!0,children:`Keep the repository private. Never include Cloudflare tokens, dashboard credentials, bypass keys, SSH keys, or machine-specific environment files.`}),(0,i.jsxs)(`p`,{children:[`If the private repository already exists, replace the final command with `,(0,i.jsx)(`code`,{children:`git remote add origin https://github.com/YOUR_ACCOUNT/code-os-skills.git`}),` and `,(0,i.jsx)(`code`,{children:`git push -u origin main`}),`.`]})]})},{title:`Install the sync worker`,children:(0,i.jsxs)(i.Fragment,{children:[(0,i.jsxs)(`p`,{children:[`Open `,(0,i.jsx)(`code`,{children:`/app/settings`}),` and set the private GitHub repository URL, `,(0,i.jsx)(`code`,{children:`~/.agents`}),` checkout, and branch. The worker reads those settings, commits local changes, rebases, and pushes. A directory lock prevents overlapping executions.`]}),(0,i.jsx)(n,{children:`mkdir -p ~/.local/bin
curl -fsSL https://code-os.mlvcdn.com/skills-sync.sh -o ~/.local/bin/code-os-skills-sync
chmod 700 ~/.local/bin/code-os-skills-sync
~/.local/bin/code-os-skills-sync`})]})},{title:`Schedule it with systemd`,children:(0,i.jsxs)(i.Fragment,{children:[(0,i.jsx)(`p`,{children:`A user timer runs the worker after boot and every two minutes. It remains independent from the Code OS dashboard process.`}),(0,i.jsx)(n,{children:`mkdir -p ~/.config/systemd/user
cat > ~/.config/systemd/user/code-os-skills-sync.service <<'EOF'
[Unit]
Description=Synchronize Code OS agent skills
After=network-online.target

[Service]
Type=oneshot
ExecStart=%h/.local/bin/code-os-skills-sync
Environment=PATH=%h/.local/bin:/usr/local/bin:/usr/bin:/bin
NoNewPrivileges=true
PrivateTmp=true
EOF

cat > ~/.config/systemd/user/code-os-skills-sync.timer <<'EOF'
[Unit]
Description=Run Code OS skills synchronization

[Timer]
OnBootSec=45s
OnUnitActiveSec=2min
Persistent=true

[Install]
WantedBy=timers.target
EOF

systemctl --user daemon-reload
systemctl --user enable --now code-os-skills-sync.timer
sudo loginctl enable-linger "$USER"`})]})},{title:`Verify the VPS side`,children:(0,i.jsxs)(i.Fragment,{children:[(0,i.jsxs)(`p`,{children:[`Trigger one execution, inspect its log, and confirm that the checkout is clean and tracking `,(0,i.jsx)(`code`,{children:`origin/main`}),`.`]}),(0,i.jsx)(n,{children:`systemctl --user start code-os-skills-sync.service
journalctl --user -u code-os-skills-sync.service -n 30 --no-pager
git -C ~/.agents status --short --branch
systemctl --user list-timers code-os-skills-sync.timer`}),(0,i.jsx)(t,{children:`A healthy run ends with “is up to date.” If Git reports a conflict, the worker stops without discarding either side.`})]})}];function o(){return(0,i.jsx)(e,{side:`VPS`,title:`Publish the VPS skill library.`,description:`Turn the existing ~/.agents directory into the private source shared by every development machine.`,steps:a,otherSide:`your computer`,otherUrl:`/skills-sync/local`})}export{o as component};