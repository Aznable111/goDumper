#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
#include <string.h>


int main(int argc, char **argv) {
	int a = 69;
	char *chungus = "my chungus";
	void *bruh = malloc(69);
	char *stuff = "MOVE ME TO THE HEAP";
	memset(bruh, 0, 69);
	memcpy(bruh, stuff, strlen(stuff) );

	pid_t pid = getpid();
	printf("running as pid: %lu\n", pid);
	while(1) {
		sleep(60);
		}
	return 0;
	
}
