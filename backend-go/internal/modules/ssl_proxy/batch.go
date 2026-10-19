// backend-go/internal/modules/ssl_proxy/batch.go
package ssl_proxy

// BatchStart 批量启动
func (s *Service) BatchStart(req *BatchReq) *BatchResp {
	return s.batch(req, func(id string) error { return s.Start(id) })
}

// BatchStop 批量停止
func (s *Service) BatchStop(req *BatchReq) *BatchResp {
	return s.batch(req, func(id string) error { return s.Stop(id) })
}

// batch 通用批量执行器（串行，避免端口/资源抢占）
func (s *Service) batch(req *BatchReq, fn func(id string) error) *BatchResp {
	resp := &BatchResp{
		Total:   len(req.ServiceIDs),
		Results: make([]BatchItemResult, 0, len(req.ServiceIDs)),
	}
	for _, sid := range req.ServiceIDs {
		r := BatchItemResult{ServiceID: sid}
		if err := fn(sid); err != nil {
			r.Success = false
			if ae, ok := err.(*APIError); ok {
				r.Code = ae.Code
				r.Message = ae.Message
			} else {
				r.Code = 50000
				r.Message = err.Error()
			}
			resp.Failed++
		} else {
			r.Success = true
			r.Code = ErrOK
			r.Message = "ok"
			resp.Succeeded++
		}
		resp.Results = append(resp.Results, r)
	}
	return resp
}
