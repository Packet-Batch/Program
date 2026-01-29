#include <stdio.h>
#include <stdlib.h>
#include <ctype.h>
#include <time.h>
#include <string.h>

#include <arpa/inet.h>

#include <helpers/int_types.h>

/**
 * Retrieves the Ethernet MAC of the host's default gateway and stores it in `mac` (u8 *).
 *
 * @param mac The variable to store the MAC address in. Must be an u8 * array with the length of ETH_ALEN (6).
 *
 * @return Void
 **/
void get_gw_mac(u8 *mac)
{
    char cmd[] = "ip neigh | grep \"$(ip -4 route list 0/0|cut -d' ' -f3) \"|cut -d' ' -f5|tr '[a-f]' '[A-F]'";

    FILE *fp = popen(cmd, "r");

    if (fp != NULL)
    {
        char line[18];

        if (fgets(line, sizeof(line), fp) != NULL)
        {
            sscanf(line, "%hhx:%hhx:%hhx:%hhx:%hhx:%hhx", &mac[0], &mac[1], &mac[2], &mac[3], &mac[4], &mac[5]);
        }

        pclose(fp);
    }
}

/**
 * Returns a random integer between min and max using rand_r(), a thread-safe function.
 *
 * @param min The minimum number to choose from.
 * @param max The maximum number to choose from.
 * @param seed The seed to pass to the rand_r() function.
 *
 * @return A 16-bit integer (u16).
 *
 * @note If you're trying to return an integer within the 8-bit range, I'd recommend casting as u8 or similar.
 **/
u16 rand_num(u16 min, u16 max, unsigned int seed)
{
    return (rand_r(&seed) % (max - min + 1)) + min;
}

/**
 * Simply lower-cases a string.
 *
 * @param str Pointer to the full string we want to lower-case.
 *
 * @return A character pointer to the lower-cased string.
 **/
char *lower_str(char *str)
{
    for (char *p = str; *p; p++)
    {
        *p = tolower(*p);
    }

    return str;
}

/**
 * Chooses a random IP from a specific CIDR range.
 *
 * @param range The range in IP/CIDR format.
 *
 * @return The pointer to a string with the random IP within the CIDR range.
 *
 * @note Thanks for the help on https://stackoverflow.com/questions/64542446/choosing-a-random-ip-from-any-specific-cidr-range-in-c.
 **/
char *rand_ip(char *range, unsigned int seed)
{
    // Split the <ip>/<cidr> and assign both values.
    char *split;

    char *s_ip = NULL;
    char *cidr_str = NULL;

    char *str = malloc(sizeof(char) * (strlen(range) + 1));
    strcpy(str, range);

    split = strtok(str, "/");

    // Check to ensure split isn't NULL.
    if (split != NULL)
    {
        // Get network IP.
        s_ip = strdup(split);

        // Split again based off of '/'.
        split = strtok(NULL, "/");

        // Check to ensure split isn't NULL and get CIDR.
        if (split != NULL)
        {
            cidr_str = strdup(split);
        }
    }

    // Check to ensure CIDR string and source IP isn't NULL.
    if (cidr_str == NULL || s_ip == NULL)
    {
        return "127.0.0.1";
    }

    // Free the temporary string (str).
    free(str);

    u8 cidr = (u8)atoi(cidr_str);

    // Create in_addr and convert the IP string to a 32-bit integer.
    struct in_addr in_addr;
    inet_aton(s_ip, &in_addr);
    u32 ip_addr = ntohl(in_addr.s_addr);

    // Get the mask (the complement of 2 to the power of the CIDR minus one).
    u32 mask = (1 << (32 - cidr)) - 1;

    // Generate a random number using rand_r(&seed).
    u32 rand_num = rand_r(&seed);

    // Generate new 32-bit IPv4 address from IP/CIDR range above.
    u32 rand_ip = (ip_addr & ~mask) | (mask & rand_num);

    // Convert the new IP to a string and print it.
    struct in_addr rand_ip_str;
    rand_ip_str.s_addr = htonl(rand_ip);

    // Free pointers we longer need.
    if (s_ip != NULL)
        free(s_ip);

    if (cidr_str != NULL)
        free(cidr_str);

    return inet_ntoa(rand_ip_str);
}

/**
 * Retrieves the source MAC address of an interface.
 *
 * @param dev The interface/device name.
 * @param src_mac A pointer to the source MAC address (u8).
 *
 * @return 0 on success or -1 on failure (path not found).
 **/
int get_src_mac_address(const char *dev, u8 *src_mac)
{
    // Format path to source MAC on file system using network class.
    char path[255];
    snprintf(path, sizeof(path) - 1, "/sys/class/net/%s/address", dev);

    // Attempt to open path/file and check.
    FILE *fp = fopen(path, "r");

    if (!fp)
    {
        return -1;
    }

    // Create buffer to copy contents of file to.
    char buffer[255];

    // Copy contents of file to buffer.
    fgets(buffer, sizeof(buffer), fp);

    // Scan MAC address.
    sscanf(buffer, "%hhx:%hhx:%hhx:%hhx:%hhx:%hhx", &src_mac[0], &src_mac[1], &src_mac[2], &src_mac[3], &src_mac[4], &src_mac[5]);

    // Close file.
    fclose(fp);

    return 0;
}