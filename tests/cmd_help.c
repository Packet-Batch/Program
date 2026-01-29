#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <linux/types.h>
#include <getopt.h>
#include <errno.h>

#include <cmd_line.h>

int main(int argc, char *argv[])
{
    print_cmd_help();

    return EXIT_SUCCESS;
}