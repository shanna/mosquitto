#ifndef _mosquitto_ext_
#define _mosquitto_ext_

#include <errno.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <mosquitto.h>

// Go external on_* callbacks.
extern void on_connect(void *);
extern void on_message(void *, char *, void *, int);
extern void on_log(void *, int, char *);

// C callbacks call external on_* Go counterpart.
void con_connect(struct mosquitto *, void *, int);
void con_message(struct mosquitto *, void *, const struct mosquitto_message *);
void con_log(struct mosquitto *, void *, int, const char *);

// Errno error messages.
void mosquitto_error(char *err);

// Wrappers for methods with boolean values. See bug: https://code.google.com/p/go/issues/detail?id=4417
struct mosquitto *mosquitto_new2(const char *, int, void *);
int mosquitto_publish2(struct mosquitto *, int *, const char *, int, const void *, int, int);

#endif

