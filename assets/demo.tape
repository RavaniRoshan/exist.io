# exist.io 90-Second Demo Walkthrough
# macOS Desktop Screen Recording Style
Output "assets/demo.gif"
Output "assets/demo.mp4"

Set Shell "bash"
Set FontSize 14
Set Width 1380
Set Height 820
Set Margin 55
Set MarginFill "assets/macos_wallpaper.png"
Set BorderRadius 14
Set WindowBar "Colorful"
Set WindowBarSize 38
Set Padding 18
Set Theme "Catppuccin Macchiato"

Hide
Type "export PATH=/home/linuxbrew/.linuxbrew/bin:/home/shiva/bin:$PATH; cd /home/shiva/projects/exist.io"
Enter
Sleep 1s
Show

Type "./scripts/demo_walkthrough.sh"
Enter
Sleep 32s
