/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package strategy

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/logger"
)

// QueryAction query target strategy object.
type QueryAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryStrategyReq
	resp *pb.QueryStrategyResp

	strategy *pbcommon.Strategy
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryStrategyReq, resp *pb.QueryStrategyResp) *QueryAction {
	action := &QueryAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *QueryAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryAction) Output() error {
	act.resp.Strategy = act.strategy
	return nil
}

func (act *QueryAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	if len(act.req.Strategyid) == 0 && (len(act.req.Appid) == 0 || len(act.req.Name) == 0) {
		return errors.New("invalid params, strategyid or appid-name missing")
	}

	if len(act.req.Strategyid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, strategyid too long")
	}

	if len(act.req.Appid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	if len(act.req.Name) > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, name too long")
	}
	return nil
}

func (act *QueryAction) query() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryStrategyReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Appid:      act.req.Appid,
		Strategyid: act.req.Strategyid,
		Name:       act.req.Name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryStrategy[%d]| request to datamanager QueryStrategy, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryStrategy(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryStrategy, %+v", err)
	}
	act.strategy = resp.Strategy

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// query strategy.
	if errCode, errMsg := act.query(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
