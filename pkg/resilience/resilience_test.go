package resilience

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	TrueOperation := func() error {
		time.Sleep(1 * time.Second)
		return nil
	}
	ErrorOperation := func() error {
		time.Sleep(1 * time.Second)
		return errors.New("some error")
	}

	CheckTrueOperation := Retry(TrueOperation, 5, 1000)
	if CheckTrueOperation != nil {
		t.Errorf("CheckTrueOperation returned %v, want nil", CheckTrueOperation)
	}

	CheckErrorOperation := Retry(ErrorOperation, 5, 1000)
	if CheckErrorOperation == nil {
		t.Errorf("CheckErrorOperation returned %v, want error", CheckErrorOperation)
	}

}

func TestTimeout(t *testing.T) {
	TrueOperation := func() error {
		return nil
	}
	ErrorOperation := func() error {
		return errors.New("some error")
	}
	ErrorTimeOperation := func() error {
		time.Sleep(10 * time.Second)
		return nil
	}
	CheckTrueOperation := Timeout(TrueOperation, 1*time.Second)
	if CheckTrueOperation != nil {
		t.Errorf("CheckTrueOperation returned %v, want nil", CheckTrueOperation)
	}
	CheckErrorOperation := Timeout(ErrorOperation, 1*time.Second)
	if CheckErrorOperation == nil {
		t.Errorf("CheckErrorOperation returned %v, want error", CheckErrorOperation)
	}
	CheckErrorTimeOperation := Timeout(ErrorTimeOperation, 1*time.Second)
	if CheckErrorTimeOperation == nil {
		t.Errorf("CheckErrorTimeOperation returned %v, want error operation timed out", CheckErrorTimeOperation)
	}
}

func TestProcessWithDLQ(t *testing.T) {
	operation := func(msg string) error {
		if msg == "msg2" || msg == "msg4" || msg == "msg6" {
			return errors.New("some error")
		}
		return nil
	}

	messagesForError := []string{"msg1", "msg2", "msg3", "msg4", "msg5", "msg6"}
	trueMessageError := []string{"msg2", "msg4", "msg6"}
	dlqForError := &[]string{}
	err := ProcessWithDLQ(messagesForError, operation, dlqForError)

	if err == nil {
		t.Errorf("ProcessWithDLQ returned nil, want error")
	} else {
		fmt.Println(*dlqForError)
		for i, v := range *dlqForError {
			if v != trueMessageError[i] {
				t.Errorf("ProcessWithDLQ returned message %v, want message %v", messagesForError[i], v)
				break
			}
		}
	}

	messagesForNil := []string{"msg1", "msg3", "msg5"}
	dlqForNil := &[]string{}
	err = ProcessWithDLQ(messagesForNil, operation, dlqForNil)
	fmt.Println(dlqForNil)
	if err != nil {
		t.Errorf("ProcessWithDLQ returned nil, want error")
	}
}
