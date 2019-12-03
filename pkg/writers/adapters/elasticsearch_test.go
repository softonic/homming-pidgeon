package adapters

import (
	"bytes"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/softonic/homing-pigeon/pkg/messages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

/**
 * Manual mock
 *
 * This is done in a manual way because the elasticsearch client does not implement
 * any interface, so we cannot mock it directly.
 */
type BulkMock struct {
	mock.Mock
}

func (b *BulkMock) getBulkFunc() esapi.Bulk {
	return func(body io.Reader, o ...func(*esapi.BulkRequest)) (*esapi.Response, error) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(body)
		args := b.Called(buf.String())
		var err error
		err = nil
		if args.Get(1) != nil {
			err = args.Get(0).(error)
		}

		return args.Get(0).(*esapi.Response), err
	}
}

func TestAdapterReceiveInvalidMessage(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	acks := esAdapter.ProcessMessages([]messages.Message{
		{
			Id:   0,
			Body: []byte("{ Invalid Json }"),
		},
	})

	bulk.AssertNotCalled(t, "func1")
	assert.Len(t, acks, 1)
	assert.False(t, acks[0].Ack)
}

func TestBulkActionWithErrorsMustDiscardAllMessages(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 404,
		Header:     nil,
		Body:       nil,
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	acks := esAdapter.ProcessMessages([]messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
	})

	bulk.AssertExpectations(t)
	assert.Len(t, acks, 1)
	assert.False(t, acks[0].Ack)
}

func TestBulkActionWithSingleItemSucessful(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       ioutil.NopCloser(strings.NewReader("{\"errors\":false,\"items\":[{\"create\":{\"status\":200}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	acks := esAdapter.ProcessMessages([]messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
	})

	bulk.AssertExpectations(t)
	assert.Len(t, acks, 1)
	assert.True(t, acks[0].Ack)
}

func TestBulkActionWithSingleItemUnsuccessful(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       ioutil.NopCloser(strings.NewReader("{\"errors\":true,\"items\":[{\"create\":{\"status\":409}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	acks := esAdapter.ProcessMessages([]messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
	})

	bulk.AssertExpectations(t)
	assert.Len(t, acks, 1)
	assert.False(t, acks[0].Ack)
}


func TestBulkActionWithMixedItemStatus(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       ioutil.NopCloser(strings.NewReader("{\"errors\":true,\"items\":[{\"create\":{\"status\":409}},{\"create\":{\"status\":200}},{\"create\":{\"status\":409}}]}")),
	}
	bulk.On("func1", mock.Anything).Once().Return(&response, nil)

	acks := esAdapter.ProcessMessages([]messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
		{
			Id:   1,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
		{
			Id:   2,
			Body: []byte("{ \"meta\": \"valid-json\" }"),
		},
	})

	bulk.AssertExpectations(t)
	assert.Len(t, acks, 3)
	assert.False(t, acks[0].Ack)
	assert.True(t, acks[1].Ack)
	assert.False(t, acks[2].Ack)
}

func TestBulkActionWithOnlyMetadata(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	response := esapi.Response{
		StatusCode: 201,
		Header:     nil,
		Body:       ioutil.NopCloser(strings.NewReader("{\"errors\":false,\"items\":[{\"delete\":{\"status\":200}}]}")),
	}
	expectedBody := "{\"delete\":{\"_id\":\"123\"}}\n"
	bulk.On("func1", expectedBody).Once().Return(&response, nil)

	acks := esAdapter.ProcessMessages([]messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"meta\": {\"delete\": {\"_id\":\"123\"}} }"),
		},
	})

	bulk.AssertExpectations(t)
	assert.Len(t, acks, 1)
	assert.True(t, acks[0].Ack)
}

func TestBulkActionWithNoMetadata(t *testing.T) {
	bulk := new(BulkMock)
	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Bulk:          bulk.getBulkFunc(),
	}

	acks := esAdapter.ProcessMessages([]messages.Message{
		{
			Id:   0,
			Body: []byte("{ \"foobar\": {\"delete\": {\"_id\":\"123\"}} }"),
		},
	})

	bulk.AssertNotCalled(t, "func1", mock.Anything)
	bulk.AssertExpectations(t)
	assert.Len(t, acks, 1)
	assert.False(t, acks[0].Ack)
}