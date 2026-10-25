package service

import (
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"go.uber.org/zap"
)

type AuditRepository interface {
	SaveAudit(audit *model.Audit) error
}

type AuditService struct {
	events          chan *model.Audit
	logger          *zap.SugaredLogger
	auditRepository AuditRepository
}

func NewAuditService(auditRepository AuditRepository, events chan *model.Audit, logger *zap.SugaredLogger) *AuditService {
	return &AuditService{
		auditRepository: auditRepository,
		events:          events,
		logger:          logger,
	}
}

func (s *AuditService) Observe() error {
	for event := range s.events {
		if err := s.auditRepository.SaveAudit(event); err != nil {
			s.logger.Errorf("save audit: %v", err)
		}
	}
	return nil
}
