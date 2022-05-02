## Linux实战技能100讲

## 什么是liunx

- 一种是Liuns编写的开源操作系统的内核
- 另一种是广义的操作系统

## liunx 版本

- 内核版本

  - http://www.kernel.org/
  - 内核版本分为三个部分
  - 主版本号、次版本号、末版本号
  - 次版本号是奇数为开发版，偶数为稳定版

- 发行版本

  - RedHat Enterprise Linux

  - Fedora

  - CentOS

  - Debian

  - Ubuntu

### 终端

- 图形终端
- 命令行
-  远程终端（SSH、VNC）

### 常用目录介绍

- `/` 根目录
- `/root`  root用户的家目录
- `/home/username`普通用户的家目录
- `/etc`配置文件目录
- `/ bin`命令目录
- `/sbin`管理命令目录
- `/usr/ bin /usr/sbin`系统预装的其他命令

### 关机

```sh
init 0
```

## 万能的帮助命令

### man 帮助

- man是manual的缩写

- man帮助用法演示.

  ```sh
  man ls
  ```

- man 也是一条命令，分为9章，可以使用man命令获得man的帮助
  ```sh
  man 1 passwd
  ```

### help 帮助

- shell(命令解释器）自带的命令称为内部命令，其他的是外部命令

- 内部命令使用help帮助
  ```sh
  help cd
  ```

- 外部命令使用help帮助

  ```sh
  ls--help
  ```
### info 帮助

info帮助比 help更详细，作为help 的补充

```sh
info ls
```

## pwd命令
pwd 显示当前的目录名称

```sh
pwd
```

## cd命令
cd 更改当前的操作目录

```sh
cd /path/to/l.... 绝对路径
cd ./path/to/... 相对路径
cd ../path/tol... 相对路径
```

```sh
cd - 返回上个目录
```



## ls命令

 ls 查看当前目录下的文件

- ls [选项，选项...]参数...

常用参数:

- -l 长格式显示文件
- -a 显示隐藏文件
- -r 逆序显示
- -t 按照时间顺序显示
- -R 递归显示

显示多个目录

```sh
ls /root /
```

## mkdir命令

创建目录

- -p 创建多级目录

```sh
mkdir a/b/c
```

## rmdir 命令

删除空目录

rm -r 删除非空目录

## cp 命令

cp复制文件和目录

- cp[选项] 文件路径
- cp[选项] 文件...路径

常用参数

- -r 复制目录
- -p 保留用户、权限、时间等文件属性
- -a 等同于-dpR

## mv命令

mv 移动文件

- mv[选项] 源文件 目标文件
- mv[选项] 源文件 目录

## 通配符

- 定义: shell 内建的符号
- 用途:操作多个相似（有简单规律)的文件
- 常用通配符

```
* 匹配任何字符串
? 匹配1个字符串
[xyz] 匹配xyz任意一个字符
[a-z] 匹配一个范围
[!xyz]或[^xyz] 不匹配
```

## 文本查看命令

- cat 文本内容显示到终端

- head 查看文件开头

  ```sh
  head -10 a.txt # 查看头10行
  ```

- tail 查看文件结尾
  - 常用参数-f 文件内容更新后，显示信息同步更新
  ```sh
  tail -10 a.txt # 查看尾10行
  ```
  
- wc 统计文件内容信息

  ```sh
  wc -l a.txt # 查看行数
  ```

## Vi/Vim 多模式文本编辑器

四种模式

- 正常模式(Normal-mode)
- 插入模式(Insert-mode)
- 命令模式(Command-mode)
- 可视模式(Visual-mode)

## 打包与压缩

### Linux的备份压缩

- 最早的Linux备份介质是磁带，使用的命令是tar
- 可以打包后的磁带文件进行压缩储存，压缩的命令是gzip和bzip2
- 经常使用的扩展名是.tar.gz   .tar.bz2 .   tgz

## 打包命令

tar 打包命令

常用参数

- c 打包
- x 解压
- f 指定操作类型为文件

**压缩**

```
tar –cvf jpg.tar *.jpg       // 将目录里所有jpg文件打包成 tar.jpg 
tar –czf jpg.tar.gz *.jpg    // 将目录里所有jpg文件打包成 jpg.tar 后，并且将其用 gzip 压缩，生成一个 gzip 压缩过的包，命名为 jpg.tar.gz 
tar –cjf jpg.tar.bz2 *.jpg   // 将目录里所有jpg文件打包成 jpg.tar 后，并且将其用 bzip2 压缩，生成一个 bzip2 压缩过的包，命名为jpg.tar.bz2 
tar –cZf jpg.tar.Z *.jpg     // 将目录里所有 jpg 文件打包成 jpg.tar 后，并且将其用 compress 压缩，生成一个 umcompress 压缩过的包，命名为jpg.tar.Z 
rar a jpg.rar *.jpg          // rar格式的压缩，需要先下载 rar for linux 
zip jpg.zip *.jpg            // zip格式的压缩，需要先下载 zip for linux
```

**解压**

```
tar –xvf file.tar         // 解压 tar 包 
tar -xzvf file.tar.gz     // 解压 tar.gz 
tar -xjvf file.tar.bz2    // 解压 tar.bz2 
tar –xZvf file.tar.Z      // 解压 tar.Z 
unrar e file.rar          // 解压 rar 
unzip file.zip            // 解压 zip 
```

**总结**

```
1、*.tar 用 tar –xvf 解压 
2、*.gz 用 gzip -d或者gunzip 解压 
3、*.tar.gz和*.tgz 用 tar –xzf 解压 
4、*.bz2 用 bzip2 -d或者用bunzip2 解压 
5、*.tar.bz2用tar –xjf 解压 
6、*.Z 用 uncompress 解压 
7、*.tar.Z 用tar –xZf 解压 
8、*.rar 用 unrar e解压 
9、*.zip 用 unzip 解压
```

## 用户与权限管理

- useradd   新建用户
- userdel    删除用户
- passwd    修改用户密码
- usermod 修改用户属性
- chage       修改用户属性

### 用户配置文件

```sh
#/etc/passwd 用户信息
#/etc/shadow 用户密码
```



## 组管理命令

- groupadd 新建用户组
- groupdel 删除用户组

## 用户切换

- su 切换用户
  - su -USERNAME 使用 login shell 方式切换用户
- sudo 以其他用户身份执行命令
  - visudo 设置需要使用sudo的用户（组）

授予user3 访问`shutdown -h` 命令

```sh
#1 
visudo

#2 添加配置
user3 ALL=/sbin/shutdown -c 
```

## 查看文件权限的方式

![image-20220502144108879](images/image-20220502144108879.png)

## 文件类型

- -普通文件
- d目录文件
- b 块特殊文件
- c 字符特殊文件
- l 符号链接
- f 命名管道
- s 套接字文件

## 文件权限的表示方式

- 字符权限表示方法
  - r 读
  - w 写
  - x 执行
- 数字权限表示方法
  - r = 4 
  - w = 2
  - x = 1 

![image-20220502144914049](images/image-20220502144914049.png)

- 创建新文件有默认权限，根据umask值计算，属主和属组根据当前进程的用户来设定

## 目录权限的表示方法

- x 进入目录
- rx 显示目录内的文件名
- wx 修改目录内的文件名







## 网络管理

- 网络状态查看
- 网络配置
- 路由命令
- 网络故障排除
- 网络服务管理
- 常用网络配置文件

## 网络状态查看

net-tools vs iproute

### 1.net-tools

- ifconfig
- route
- netstat

### 2.iproute2

- ip
- ss

## ifconfig 

- etho第一块网卡（网络接口)
- 你的第一个网络接口可能叫做下面的名字. eno1板载网卡.
  - ens33 PCl-E网卡
  - enp0s3无法获取物理信息的PCI-E网卡
  - centOS 7使用了一致性网络设备命名，以上都不匹配则使用etho

## 网络接口命名修改

- 网卡命名规则受biosdevname和net.ifnames两个参数影响

- 编辑/etc/default/grub文件，增加biosdevname=0 net.ifnames=0

- 更新grub

   grub2-mkconfig -o /boot/grub2/grub.cfg


- 重启

  reboot

  |       | biosdevname | net.ifnames | 网卡名 |
  | :---: | :---------: | :---------: | :----: |
  | 默认  |      0      |      1      | ens33  |
  | 组合1 |      1      |      0      |  em1   |
  | 组合2 |      0      |      0      |  eth0  |



## 查看网卡物理连接情况

- mii-tool eth0

## 查看网关

- route -n
- 使用-n参数不解析主机名

 
