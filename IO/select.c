//服务端
#include <stdio.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>
#include <sys/select.h>
int main() {
    // 创建socket
    int lfd = socket(PF_INET, SOCK_STREAM, 0);
    struct sockaddr_in saddr;
    saddr.sin_port = htons(9999);
    saddr.sin_family = AF_INET;
    saddr.sin_addr.s_addr = INADDR_ANY;
    // 绑定
    bind(lfd, (struct sockaddr *)&saddr, sizeof(saddr));
    // 监听
    listen(lfd, 8);
    // 创建一个fd_set的集合，存放的是需要检测的文件描述符
    fd_set rdset, tmp;	//fd_set底层可以表示1024个文件描述符
    FD_ZERO(&rdset);	//初始化
    FD_SET(lfd, &rdset);	//添加需要监听的文件描述符
    int maxfd = lfd;	//定义最大文件描述符，作为参数传入select函数中
    while(1) {
        tmp = rdset;	//rdset这个不能变，因为内核再检测时，如果没有数据，就会将其变为0，因此，我们需要复制一份。
        // 调用select系统函数，让内核帮检测哪些文件描述符有数据
        int ret = select(maxfd + 1, &tmp, NULL, NULL, NULL);
        if(ret == -1) {
            perror("select");
            exit(-1);
        } else if(ret == 0) {	//这里不可能为0，因为设置了永久阻塞NULL，直到检测到文件描述符有数据变化
            continue;	
        } else if(ret > 0) {	//ret只会返回文件描述符发生变化的个数，不知道具体哪个发生了变化，需要遍历查找
            // 说明检测到了有文件描述符的对应的缓冲区的数据发生了改变
            if(FD_ISSET(lfd, &tmp)) {	//lfd为监听文件描述符
                // 表示有新的客户端连接进来了
                struct sockaddr_in cliaddr;
                int len = sizeof(cliaddr);
                int cfd = accept(lfd, (struct sockaddr *)&cliaddr, &len);
                // 将新的文件描述符加入到集合中，下一次select检测时，需要检测这些通信的文件描述符有没有数据
                FD_SET(cfd, &rdset);
                // 更新最大的文件描述符
                maxfd = maxfd > cfd ? maxfd : cfd;
            }
            //检测剩余文件描述符有没有数据变化，从lfd+1开始即可
            for(int i = lfd + 1; i <= maxfd; i++) {
                if(FD_ISSET(i, &tmp)) {
                    // 说明这个文件描述符对应的客户端发来了数据
                    char buf[1024] = {0};
                    int len = read(i, buf, sizeof(buf));
                    if(len == -1) {
                        perror("read");
                        exit(-1);
                    } else if(len == 0) {	//说明客户端断开连接
                        printf("client closed...");
                        close(i);	//关闭文件描述符
                        FD_CLR(i, &rdset);	//fd_set中不在监测这个文件描述符
                    } else if(len > 0) {
                        printf("read buf = %s", buf);
                        write(i, buf, strlen(buf) + 1);
                    }
                }
            }
        }
    }
    close(lfd);
    return 0;
}