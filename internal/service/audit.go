package service

import (
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"go.uber.org/zap"
)

// AuditRepository сохраняет события аудита.
type AuditRepository interface {
	// SaveAudit записывает одно событие аудита.
	SaveAudit(audit *model.Audit) error
}

// AuditService читает события аудита из канала и передаёт их в хранилище.
type AuditService struct {
	events          chan *model.Audit
	logger          *zap.SugaredLogger
	auditRepository AuditRepository
}

// NewAuditService создаёт сервис, который слушает events и пишет их через auditRepository.
func NewAuditService(auditRepository AuditRepository, events chan *model.Audit, logger *zap.SugaredLogger) *AuditService {
	return &AuditService{
		auditRepository: auditRepository,
		events:          events,
		logger:          logger,
	}
}

// Observe читает канал событий до его закрытия и сохраняет каждое событие.
// Ошибка записи логируется и не останавливает цикл. После закрытия канала Observe возвращает nil.
func (s *AuditService) Observe() error {
	for event := range s.events {
		if err := s.auditRepository.SaveAudit(event); err != nil {
			s.logger.Errorf("save audit: %v", err)
		}
	}
	return nil
}
