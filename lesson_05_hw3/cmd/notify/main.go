package main

import (
	"errors"
	"fmt"

	"lesson5/student/practice"
)

func main() {
	email := newEmail()
	push := newPush("phone-1")

	successfulDispatch(email, push)
	emptyMessage(email, push)
	noDeviceID()
	brokenChannel()
	smsChannel()
}

// Успешный Dispatch
func successfulDispatch(email *practice.EmailNotifier, push *practice.PushNotifier) {
	title(1, "Успех: Dispatch(email, push, \"Текст уведомления\")")
	err := practice.Dispatch([]practice.Notifier{email, push}, "Текст уведомления")

	report(err)
	history("email", email.Messages)
	history("push", push.Messages)
}

// Пустой текст
func emptyMessage(email *practice.EmailNotifier, push *practice.PushNotifier) {
	title(2, "Пустой текст: Dispatch(email, push, \"\")")
	err := practice.Dispatch([]practice.Notifier{email, push}, "")

	report(err)
	history("email", email.Messages)
	history("push", push.Messages)
}

// Нет настроек
func noDeviceID() {
	title(3, "Нет настройки: push без DeviceID")

	push := newPush("")
	err := practice.Dispatch([]practice.Notifier{push}, "Текст уведомления")

	report(err)
	history("push", push.Messages)
}

// Паника одного канала не мешает последующим
func brokenChannel() {
	title(4, "Сломанная реализация: SafeDispatch(broken, push)")

	broken := &practice.BrokenNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "broken"},
	}

	push := newPush("phone-2")

	reportAll(practice.SafeDispatch([]practice.Notifier{broken, push}, "Текст уведомления"))

	history("push", push.Messages)
}

// Проверка, что SMS канал работает без изменений диспетчеров
func smsChannel() {
	title(5, "Дополнительно: SMS через SafeDispatch")

	sms := &practice.SMSNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "sms"},
		Phone:        "+7-000-000-00-00",
	}

	noPhone := &practice.SMSNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "sms-backup"},
	}

	reportAll(practice.SafeDispatch([]practice.Notifier{noPhone, sms}, "Текст уведомления"))

	history("sms", sms.Messages)
}

// report выбирает текст по причине ошибки
func report(err error) {
	var sendErr *practice.SendError

	switch {
	case nil == err:
		fmt.Println("ошибок нет")
	case errors.Is(err, practice.ErrEmptyMessage):
		fmt.Println("пустой текст:", err)
	case errors.As(err, &sendErr):
		fmt.Printf("ошибка настройки: канал %q, причина %q\n", sendErr.Channel, sendErr.Reason)
	case errors.Is(err, practice.ErrNotifierPanic):
		fmt.Println("перехватили панику:", err)
	default:
		fmt.Println("неизвестная ошибка:", err)
	}
}

// reportAll печатает все ошибки SafeDispatch: errors.Is принимает одну ошибку
func reportAll(errs []error) {
	if 0 == len(errs) {
		report(nil)
	}

	for _, err := range errs {
		report(err)
	}
}

func title(n int, text string) {
	fmt.Printf("\n%d. %s\n", n, text)
}

func history(channel string, messages []string) {
	fmt.Printf("история %s: %q\n", channel, messages)
}

func newEmail() *practice.EmailNotifier {
	return &practice.EmailNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "email"},
		Address:      "student@example.com",
	}
}

func newPush(deviceID string) *practice.PushNotifier {
	return &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:     deviceID,
	}
}
