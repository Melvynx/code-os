import { createFileRoute } from '@tanstack/react-router'
import { SyncCode, SyncGuideLayout, SyncNotice } from '../components/skills-sync-docs'

export const Route = createFileRoute('/skills-sync_/local')({
  head: () => ({
    meta: [
      { title: 'Computer Skills Sync Setup — Code OS' },
      { name: 'description', content: 'Connect a Mac or local development computer to the shared private agent skill library.' },
    ],
    links: [{ rel: 'canonical', href: 'https://code-os.mlvcdn.com/skills-sync/local' }],
  }),
  component: LocalSkillsSyncPage,
})

const steps = [
  {
    title: 'Prepare Git authentication',
    children: <><p>Use GitHub CLI over HTTPS so the LaunchAgent can authenticate through the macOS Keychain without depending on an interactive SSH agent.</p><SyncCode>{`gh auth status || gh auth login
gh config set git_protocol https
gh auth setup-git
git config --global user.name "YOUR_NAME"
git config --global user.email "YOUR_EMAIL"`}</SyncCode></>,
  },
  {
    title: 'Preserve the local library',
    children: <><p>Quit Codex, Cursor, and any agent currently reading skills. Move the existing local directory aside, then clone the VPS-backed repository into the canonical <code>~/.agents</code> path.</p><SyncCode>{`backup_dir="$HOME/.agents.backup.$(date +%Y%m%d-%H%M%S)"
if [ -d ~/.agents ]; then mv ~/.agents "$backup_dir"; fi
git clone https://github.com/YOUR_ACCOUNT/code-os-skills.git ~/.agents
echo "Previous local skills: $backup_dir"`}</SyncCode><SyncNotice>Do not delete the backup. Compare it with the cloned library and manually copy only the local skills or rules you still need.</SyncNotice></>,
  },
  {
    title: 'Install the same worker',
    children: <><p>Configure the same repository, checkout, and branch in the local Code OS dashboard. Both machines then use the identical audited command and private Git branch. Run it once interactively before scheduling it.</p><SyncCode>{`mkdir -p ~/.local/bin
curl -fsSL https://code-os.mlvcdn.com/skills-sync.sh -o ~/.local/bin/code-os-skills-sync
chmod 700 ~/.local/bin/code-os-skills-sync
~/.local/bin/code-os-skills-sync`}</SyncCode></>,
  },
  {
    title: 'Schedule it with launchd',
    children: <><p>The LaunchAgent starts after login and runs every two minutes. Logs stay in <code>~/Library/Logs/CodeOS</code>.</p><SyncCode>{`mkdir -p ~/Library/LaunchAgents ~/Library/Logs/CodeOS
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
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/dev.code-os.skills-sync.plist`}</SyncCode></>,
  },
  {
    title: 'Verify both directions',
    children: <><p>Create a harmless test file on the computer, run a sync, then confirm it arrives on the VPS after the next timer execution.</p><SyncCode>{`touch ~/.agents/.sync-check
~/.local/bin/code-os-skills-sync
launchctl kickstart -k gui/$(id -u)/dev.code-os.skills-sync
tail -n 30 ~/Library/Logs/CodeOS/code-os-skills-sync.log
git -C ~/.agents status --short --branch`}</SyncCode><p>On the VPS, run <code>systemctl --user start code-os-skills-sync.service</code>, check that <code>.sync-check</code> exists, then remove it and sync once more.</p><SyncNotice security>Only <code>~/.agents</code> is synchronized. Project repositories, Code OS configuration, tokens, and screenshot bypass keys stay machine-local.</SyncNotice></>,
  },
]

function LocalSkillsSyncPage() {
  return <SyncGuideLayout side="YOUR COMPUTER · macOS" title="Connect your local skill library." description="Clone the VPS-backed source, preserve any local-only work, and keep both sides synchronized in the background." steps={steps} otherSide="the VPS" otherUrl="/skills-sync/vps" />
}
