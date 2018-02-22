package cloudfoundry_test

import (
	"errors"

	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-spinner/cloudfoundry"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-spinner/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Emitter", func() {
	It("logs to stdout once", func() {
		mockWriter := fakes.Writer{}

		writer := cloudfoundry.NewEmitter(&mockWriter, 1, 0)
		writer.Emit("something")

		Expect(mockWriter.Writes).To(HaveLen(1))
		Expect(mockWriter.Writes[0]).To(ContainSubstring("something"))
	})

	It("logs to stdout x specified times", func() {
		mockWriter := fakes.Writer{}

		writer := cloudfoundry.NewEmitter(&mockWriter, 10, 0)
		writer.Emit("something")

		Expect(mockWriter.Writes).To(HaveLen(10))
	})

	It("returns a count of successfully emitted logs", func() {
		mockWriter := fakes.Writer{}

		writer := cloudfoundry.NewEmitter(&mockWriter, 10, 0)
		count, _ := writer.Emit("something")

		Expect(count).To(Equal(10))
	})

	It("returns zero when no logs are emitted", func() {
		mockWriter := fakes.FailingWriter{}
		mockWriter.Err = errors.New("Fail!!")

		writer := cloudfoundry.NewEmitter(&mockWriter, 10, 0)
		count, _ := writer.Emit("something")

		Expect(count).To(Equal(0))
	})
})
