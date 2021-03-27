/*
Package grq implements persistent, thread and cross process safe task queue, that uses redis as backend.
It should be used, when RabbitMQ is too complicated, and MQTT is not enough (because it cannot cache messages),
and BeanStalkd is classic from 21 september of 2007 year, that is hard to find in many linux distros.

GRQ mean `Golang Redis Queue`.
*/
package grq
