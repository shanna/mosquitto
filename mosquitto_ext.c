#include "mosquitto_ext.h"

void con_connect(struct mosquitto *mosq, void *userdata, int result) {
  if (!result) on_connect(userdata);
}

void con_message(struct mosquitto *mosq, void *userdata, const struct mosquitto_message *message) {
  on_message(userdata, message->topic, message->payload, message->payloadlen);
}

void con_log(struct mosquitto *mosq, void *userdata, int level, const char *str) {
  on_log(userdata, level, (char *)str);
}

void mosquitto_error(char *err) {
#ifndef WIN32
  strerror_r(errno, err, 1024);
#else
  FormatMessage(FORMAT_MESSAGE_FROM_SYSTEM, NULL, errno, 0, (LPTSTR)err, 1024, NULL);
#endif
}

struct mosquitto *mosquitto_new2(const char *id, int clean_session, void *userdata) {
  struct mosquitto *mosq = mosquitto_new(id, clean_session, userdata);
  if (mosq) {
    mosquitto_connect_callback_set(mosq, con_connect);
    mosquitto_message_callback_set(mosq, con_message);
    mosquitto_log_callback_set(mosq, con_log);
  }
  return mosq;
}

int mosquitto_publish2(struct mosquitto *mosq, int *mid, const char *topic, int payloadlen, const void *payload, int qos, int retain) {
  return mosquitto_publish(mosq, mid, topic, payloadlen, payload, qos, retain);
}

