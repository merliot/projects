    1  uptime
    2  sudo apt update
    3  git
    4  sudo apt install git -y
    5  https://github.com/merliot/projects.git
    6  git clone --depth=1 https://github.com/merliot/projects.git
    7  cd projects/garden/
    8  go
    9  wget https://go.dev/dl/go1.19.linux-armv6l.tar.gz
   10  ls
   11  ls /usr/local/go
   12  sudo tar -C /usr/local -zxf go1.19.linux-armv6l.tar.gz 
   13  export PATH=/usr/local/go/bin:$PATH
   14  echo $PATH
   15  go version
   16  ./build
   17  ~/go/bin/garden 
   18  ./auto-start 
   19  vi ../scripts/ap-start 
   20  ../scripts/ap-start 
   21  sudo reboot
   22  sudo raspi-config
   23  l
   24  ls
   25  ls /boot
   26  vi /boot/config.txt 
   27  sudo raspi-config -h
   28  sudo /etc/init.d/ssh stop
   29  sudo update-rc.d ssh disable
   30  sudo raspi-config 
   31  history
