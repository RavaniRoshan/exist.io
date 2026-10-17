# exist.io 90-Second Demo Walkthrough
# Neutral, self-hosted Git forge built for coding agents
Output "assets/demo.gif"
Output "assets/demo.mp4"

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 720
Set Padding 20
Set Theme "Catppuccin Macchiato"

Hide
Type "export PATH=/home/linuxbrew/.linuxbrew/bin:/home/shiva/bin:$PATH; cd /home/shiva/projects/exist.io"
Enter
Sleep 1s
Show

Type "./scripts/demo_walkthrough.sh"
Enter
Sleep 32s
