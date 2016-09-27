package sink

type StackdriverClient interface {
	Post(payload interface{})
}


/*


3	type MockSplunkClient struct {
4		CapturedEvents []map[string]interface{}
5		PostBatchFn    func(events []map[string]interface{}) error
6	}
7
8	func (m *MockSplunkClient) Post(events []map[string]interface{}) error {
 */