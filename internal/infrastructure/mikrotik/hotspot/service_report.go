package hotspot

// ========== REPORTS ==========

// GetReportByDate mendapatkan report berdasarkan tanggal
func (s *Service) GetReportByDate(date string, useCache bool) ([]map[string]string, error) {
	reply, err := s.client.Run("/system/script/print", "?source="+date)
	if err != nil {
		return nil, err
	}

	reports := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		reports[i] = re.Map
	}

	return reports, nil
}

// GetReportCount mendapatkan jumlah report
func (s *Service) GetReportCount(date string) (int, error) {
	reply, err := s.client.Run("/system/script/print", "?source="+date)
	if err != nil {
		return 0, err
	}

	return len(reply.Re), nil
}