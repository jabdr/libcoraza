#include <stdio.h>

#include "coraza/coraza.h"

void logcb(const char *data)
{
    printf("%s\n", data);
}


int main()
{
    coraza_config_t *cfg = NULL;
    coraza_waf_t *waf = NULL;
    coraza_transaction_t *tx = NULL;
    coraza_intervention_t *intervention = NULL;
    char *err = NULL;

    printf("Starting...\n");
    cfg = coraza_new_config();
    printf("Attaching log callback\n");
    coraza_set_log_cb(waf, logcb);

    printf("Compiling rules...\n");
    coraza_rules_add(cfg, "SecRule REMOTE_ADDR \"127.0.0.1\" \"id:1,phase:1,deny,log,msg:'test 123',status:403\"");
    if(err) {
        printf("%s\n", err);
        return 1;
    }
    
    waf = coraza_new_waf(cfg, &err);
    if (waf == 0) {
        printf("Failed to create waf: %s\n", err);
        free(err);
        return 1;
    }

    printf("%d rules compiled\n", coraza_rules_count(waf));
    printf("Creating transaction...\n");
    tx = coraza_new_transaction(waf, NULL);
    if(tx == 0) {
        printf("Failed to create transaction\n");
        return 1;
    }

    printf("Processing connection...\n");
    coraza_process_connection(tx, "127.0.0.1", 55555, "", 80);
    printf("Processing request line\n");
    coraza_process_uri(tx, "/someurl", "GET", "HTTP/1.1");
    printf("Processing phase 1\n");
    coraza_process_request_headers(tx);
    printf("Processing phase 2\n");
    coraza_process_request_body(tx);
    printf("Processing phase 3\n");
    coraza_process_response_headers(tx, 200, "HTTP/1.1");
    printf("Processing phase 4\n");
    coraza_process_response_body(tx);
    printf("Processing phase 5\n");
    coraza_process_logging(tx);
    printf("Processing intervention\n");

    intervention = coraza_intervention(tx);
    if (intervention == NULL)
    {
        printf("Failed to disrupt transaction\n");
        return 1;
    }
    printf("Transaction disrupted with status %d\n", intervention->status);

    if(coraza_free_transaction(tx, &err) != 0) {
        printf("Failed to free transaction 1 %s\n", err);
        free(err);
        return 1;
    }
    coraza_free_intervention(intervention);
    coraza_free_waf(waf);
    coraza_free_config(cfg);
    return 0;
}
