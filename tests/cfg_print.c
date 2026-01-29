#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <linux/types.h>
#include <getopt.h>
#include <errno.h>

#include <cmd_line.h>
#include <config.h>

int main(int argc, char *argv[])
{
    struct cmd_line cmd = {0};

    // Parse command line and store values into cmd.
    parse_cli(argc, argv, &cmd);

    // Check if config is specified.
    if (cmd.config == NULL)
    {
        // Copy default values.
        cmd.config = "./data/conf.json";

        // Let us know if we're using the default config when the verbose flag is specified.
        if (cmd.verbose)
        {
            fprintf(stdout, "No config specified. Using default: %s.\n", cmd.config);
        }
    }

    // Create config structure.
    struct config *cfg = malloc(sizeof(struct config));
    memset(cfg, 0, sizeof(*cfg));

    int seq_cnt = 0;

    // Set default values on each sequence.
    for (int i = 0; i < MAX_SEQUENCES; i++)
    {
        clear_sequence(cfg, i);
    }

    // Attempt to parse config.
    parse_config(cmd.config, cfg, 0, &seq_cnt, 0);

    print_config(cfg, seq_cnt);

    free(cfg);

    return EXIT_SUCCESS;
}