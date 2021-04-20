package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// DirektivURL stroes the endpoint for direktiv
var DirektivURL string

// GenerateCmd is a basic cobra function
func GenerateCmd(use, short, long string, fn func(cmd *cobra.Command, args []string), c cobra.PositionalArgs) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run:   fn,
		Args:  c,
	}
}

func DoRequest(method, path string) ([]byte, error) {

	var ret []byte

	d := time.Now().Add(10 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method,
		fmt.Sprintf("%s/api/%s", DirektivURL, path), nil)
	if err != nil {
		return ret, err
	}

	c := &http.Client{}

	res, err := c.Do(req)
	if err != nil {
		return ret, err
	}
	defer res.Body.Close()

	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ret, err
	}

	if res.StatusCode != 200 {

	}

	return out, nil

}

// func CreateClient(conn *grpc.ClientConn) (ingress.DirektivIngressClient, context.Context, context.CancelFunc) {
// 	client := ingress.NewDirektivIngressClient(conn)
//
// 	// set context with 3 second timeout
// 	ctx := context.Background()
// 	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*3))
//
// 	cancelConns := func() {
// 		conn.Close()
// 		cancel()
// 	}
//
// 	return client, ctx, cancelConns
// }
//
// // RequestObject is the json output for most commands that return a string.
// type RequestObject struct {
// 	Message    string `json:"message"`
// 	Successful bool   `json:"success"`
// }
//
// // WriteRequestJSON writes the entire output of an object with indentation
// func WriteRequestJSON(message string, success bool, logger elog.View) {
// 	r := &RequestObject{
// 		Message:    message,
// 		Successful: success,
// 	}
// 	data, err := json.MarshalIndent(r, "", "    ")
// 	if err != nil {
// 		r.Successful = false
// 		r.Message = err.Error()
// 		logger.Printf("%s", data)
// 		os.Exit(1)
// 	}
// 	logger.Printf("%s", data)
// }
//
// // WriteJSON writes the entire output of an object with indentation
// func WriteJSON(data interface{}, logger elog.View) {
// 	bv, err := json.MarshalIndent(data, "", "    ")
// 	if err != nil {
// 		logger.Errorf("%s", err.Error())
// 		os.Exit(1)
// 	}
// 	logger.Printf("%s", bv)
// }
//
// // List returns a list of namespaces, workflows or instances as json output. Returns null list if length is 0
// type List struct {
// 	List interface{} `json:"list,omitempty"`
// }
//
// func WriteJsonList(list interface{}, logger elog.View) {
// 	listObj := &List{
// 		List: list,
// 	}
// 	data, err := json.MarshalIndent(listObj, "", "    ")
// 	if err != nil {
// 		logger.Errorf("%s", err.Error())
// 		os.Exit(1)
// 	}
// 	logger.Printf("%s", data)
// }
