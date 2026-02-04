#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <getopt.h>
#include <errno.h>
#include <signal.h>
#include <sys/resource.h>

#include <constants.h>

#include <common/utils.h>
#include <common/cli.h>
#include <common/config.h>

#include <tech/sequence.h>

#ifdef ENABLE_AF_XDP
#include <tech/af-xdp/af_xdp.h>
#endif

extern int errno;

config_t *cfg = NULL;

/**
 * Signal handler to shut down the program.
 *
 * @return Void
 */
void sign_hdl(int sig)
{
    sequence__stop_all(cfg);

    exit(EXIT_SUCCESS);
}

/**
 * The main program.
 *
 * @param argc The amount of arguments.
 * @param argv An array of arguments passed to program.
 *
 * @return Int (exit code)
 */
int main(int argc, char **argv)
{
    // Create command line structure.
    opterr = 0;
    cli_t cmd = {0};

    cmd.tech_af_xdp.batch_size = 64;

    // Parse command line and store values into cmd.
    cli__parse(argc, argv, &cmd);

    // Help menu.
    if (cmd.help)
    {
        cli__print_help();

        return EXIT_SUCCESS;
    }

    // Set global variables in AF_XDP program.
    tech_afxdp__setup_vars(&cmd.tech_af_xdp, cmd.verbose);

    // Check if config is specified.
    if (cmd.config == NULL)
    {
        // Copy default values.
        cmd.config = CONF_PATH_DEFAULT;

        // Let us know if we're using the default config when the verbose flag is specified.
        if (cmd.verbose)
        {
            fprintf(stdout, "No config specified. Using default: %s.\n", cmd.config);
        }
    }

    // Allocate memory for config.
    // We allocate on heap due to size of config.
    cfg = malloc(sizeof(config_t));
    memset(cfg, 0, sizeof(*cfg));

    int seq_cnt = 0;

    // We need to setup default values for all sequences before starting.
    for (int i = 0; i < MAX_SEQUENCES; i++)
    {
        config__clr_seq(cfg, i);
    }

    // Attempt to parse config and we need to determine whether we're using CLI first sequence override.
    u8 log = 1;

    if (cmd.cli)
    {
        fprintf(stdout, "Using CLI first sequence override...\n");
        log = 0;
    }

    config__parse(cmd.config, cfg, 0, &seq_cnt, log);

    // If CLI is specified, parse sequence options.
    if (cmd.cli)
    {
        cli__parse_seq_opts(&cmd, cfg);

        // Ensure we have at least one sequence so the program doesn't immediately close.
        if (seq_cnt < 1)
            seq_cnt = 1;
    }

    // Check for list option. If so, print helpful information for configuration.
    if (cmd.list)
    {
        config__print(cfg, seq_cnt);

        return EXIT_SUCCESS;
    }

#ifdef CONF_UNLOCK_RLIMIT
    // Raising the stack limit to unlimited if needed.
    struct rlimit rl;
    rl.rlim_cur = RLIM_INFINITY;
    rl.rlim_max = RLIM_INFINITY;

    if (setrlimit(RLIMIT_STACK, &rl) != 0)
    {
        fprintf(stderr, "Failed to set stack limit to unlimited\n");
    }
#endif

    // Setup signals so we can catch and gracefully shutdown the program instead of killing it.
    signal(SIGINT, sign_hdl);
    signal(SIGTERM, sign_hdl);

    // Loop through all sequences and start them.
    for (int i = 0; i < seq_cnt; i++)
    {
        sequence__start(cfg->interface, cfg->seq[i], seq_cnt, cmd, cmd.tech_af_xdp.batch_size);

        // Sleep for 1 second between sequences.
        sleep(1);
    }

    // Stop all sequences and print stats.
    sequence__stop_all(cfg);

    // Close program successfully.
    return EXIT_SUCCESS;
}