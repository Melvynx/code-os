import{n as e,r as t,t as n}from"./skills-sync-docs-D0By4tkc.js";import{i as r}from"./index-CZop_SAX.js";var i=r(),a=[{title:`Prepare Git authentication`,children:(0,i.jsxs)(i.Fragment,{children:[(0,i.jsx)(`p`,{children:`Use GitHub CLI over HTTPS so the LaunchAgent can authenticate through the macOS Keychain without depending on an interactive SSH agent.`}),(0,i.jsx)(n,{children:`gh auth status || gh auth login
gh config set git_protocol https
gh auth setup-git
git config --global user.name "YOUR_NAME"
git config --global user.email "YOUR_EMAIL"`})]})},{title:`Preserve the local library`,children:(0,i.jsxs)(i.Fragment,{children:[(0,i.jsxs)(`p`,{children:[`Quit Codex, Cursor, and any agent currently reading skills. Move the existing local directory aside, then clone the VPS-backed repository into the canonical `,(0,i.jsx)(`code`,{children:`~/.agents`}),` path.`]}),(0,i.jsx)(n,{children:`backup_dir="$HOME/.agents.backup.$(date +%Y%m%d-%H%M%S)"
if [ -d ~/.agents ]; then mv ~/.agents "$backup_dir"; fi
git clone https://github.com/YOUR_ACCOUNT/code-os-skills.git ~/.agents
echo "Previous local skills: $backup_dir"`}),(0,i.jsx)(t,{children:`Do not delete the backup. Compare it with the cloned library and manually copy only the local skills or rules you still need.`})]})},{title:`Install the same worker`,children:(0,i.jsxs)(i.Fragment,{children:[(0,i.jsx)(`p`,{children:`Configure the same repository, checkout, and branch in the local Code OS dashboard. Both machines then use the identical audited command and private Git branch. Run it once interactively before scheduling it.`}),(0,i.jsx)(n,{children:`mkdir -p ~/.local/bin
curl -fsSL https://code-os.mlvcdn.com/skills-sync.sh -o ~/.local/bin/code-os-skills-sync
chmod 700 ~/.local/bin/code-os-skills-sync
~/.local/bin/code-os-skills-sync`})]})},{title:`Schedule it with launchd`,children:(0,i.jsxs)(i.Fragment,{children:[(0,i.jsxs)(`p`,{children:[`The LaunchAgent starts after login and runs every two minutes. Logs stay in `,(0,i.jsx)(`code`,{children:`~/Library/Logs/CodeOS`}),`.`]}),(0,i.jsx)(n,{children:`mkdir -p ~/Library/LaunchAgents ~/Library/Logs/CodeOS
cat > ~/Library/LaunchAgents/dev.code-os.skills-sync.plist <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>dev.code-os.skills-sync</string>
  <key>ProgramArguments</key><array>
    <string>/bin/zsh</string><string>-lc</string>
    <string>$HOME/.local/bin/code-os-skills-sync</string>
  </array>
  <key>RunAtLoad</key><true/>
  <key>StartInterval</key><integer>120</integer>
  <key>StandardOutPath</key><string>LOCAL_LOG_PATH/code-os-skills-sync.log</string>
  <key>StandardErrorPath</key><string>LOCAL_LOG_PATH/code-os-skills-sync.error.log</string>
</dict></plist>
EOF
sed -i '' "s|LOCAL_LOG_PATH|$HOME/Library/Logs/CodeOS|g" ~/Library/LaunchAgents/dev.code-os.skills-sync.plist
launchctl bootout gui/$(id -u) ~/Library/LaunchAgents/dev.code-os.skills-sync.plist 2>/dev/null || true
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/dev.code-os.skills-sync.plist`})]})},{title:`Verify both directions`,children:(0,i.jsxs)(i.Fragment,{children:[(0,i.jsx)(`p`,{children:`Create a harmless test file on the computer, run a sync, then confirm it arrives on the VPS after the next timer execution.`}),(0,i.jsx)(n,{children:`touch ~/.agents/.sync-check
~/.local/bin/code-os-skills-sync
launchctl kickstart -k gui/$(id -u)/dev.code-os.skills-sync
tail -n 30 ~/Library/Logs/CodeOS/code-os-skills-sync.log
git -C ~/.agents status --short --branch`}),(0,i.jsxs)(`p`,{children:[`On the VPS, run `,(0,i.jsx)(`code`,{children:`systemctl --user start code-os-skills-sync.service`}),`, check that `,(0,i.jsx)(`code`,{children:`.sync-check`}),` exists, then remove it and sync once more.`]}),(0,i.jsxs)(t,{security:!0,children:[`Only `,(0,i.jsx)(`code`,{children:`~/.agents`}),` is synchronized. Project repositories, Code OS configuration, tokens, and screenshot bypass keys stay machine-local.`]})]})}];function o(){return(0,i.jsx)(e,{side:`YOUR COMPUTER · macOS`,title:`Connect your local skill library.`,description:`Clone the VPS-backed source, preserve any local-only work, and keep both sides synchronized in the background.`,steps:a,otherSide:`the VPS`,otherUrl:`/skills-sync/vps`})}export{o as component};