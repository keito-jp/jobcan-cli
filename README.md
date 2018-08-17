# Jobcan CLI

## Install

```shell
go get github.com/keito-jp/jobcan-cli
```

## Usage

```shell
jobcan-cli
```

## Environment Variables

Name                     |Description
-------------------------|--------------------------
`JOBCAN_CLIENT_ID`       |Client ID of Jobcan
`JOBCAN_EMAIL`           |Email Address of Jobcan
`JOBCAN_PASSWORD`        |Password of Jobcan
`JOBCAN_SLACK_API_TOKEN` |API Token of Slack
`JOBCAN_SLACK_NAME`      |Display Name of Slack Post

## Package jobcan
--
    import "github.com/syoya/slack-button/jobcan"

Package jobcan provides interfaces that enable to use at go codes.

### Usage

##### type Error

```go
type Error struct {
	Message string
	Status  string
}
```

Error is Error type of Jobcan class.

##### func (*Error) Error

```go
func (err *Error) Error() string
```

##### type Jobcan

```go
type Jobcan struct {
}
```

Jobcan is the struct for defining the Jobcan class.

##### func  NewJobcan

```go
func NewJobcan(clientID string, email string, password string) (*Jobcan, error)
```
NewJobcan is constructor of Jobcan class.

##### func (*Jobcan) Punch

```go
func (j *Jobcan) Punch() error
```
Punch punch in

##### func (*Jobcan) Status

```go
func (j *Jobcan) Status() (string, error)
```
Status read current status of Jobcan.

##### type Kintai

```go
type Kintai struct {
	Result        int          `json:"result"`
	State         int          `json:"state"`
	CurrentStatus string       `json:"current_status"`
	Errors        KintaiErrors `json:"errors"`
}
```

Kintai is the struct for save result of punching in.

##### type KintaiErrors

```go
type KintaiErrors struct {
	AditCount string `json:"aditCount"`
}
```

KintaiErrors is the struct for error of punching in.
