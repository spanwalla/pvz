package service_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/spanwalla/pvz/internal/dto"
	"github.com/spanwalla/pvz/internal/entity"
	metricmocks "github.com/spanwalla/pvz/internal/metrics/mocks"
	"github.com/spanwalla/pvz/internal/repository"
	repomocks "github.com/spanwalla/pvz/internal/repository/mocks"
	"github.com/spanwalla/pvz/internal/service"
)

func TestPointService_Create(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		arbitraryErr = errors.New("arbitrary error")
		ctx          = context.Background()
		city         = "Екатеринбург"
		timestamp    = time.Now()
	)

	point := entity.Point{
		ID:        uuid.New(),
		City:      city,
		CreatedAt: timestamp,
	}

	type MockBehavior func(p *repomocks.MockPoint, m *metricmocks.MockCounter)

	for _, tc := range []struct {
		name         string
		mockBehavior MockBehavior
		want         entity.Point
		wantErr      error
	}{
		{
			name: "success",
			mockBehavior: func(p *repomocks.MockPoint, m *metricmocks.MockCounter) {
				p.EXPECT().Create(ctx, city).Return(point, nil)
				m.EXPECT().Inc()
			},
			want: point,
		},
		{
			name: "city not found",
			mockBehavior: func(p *repomocks.MockPoint, m *metricmocks.MockCounter) {
				p.EXPECT().Create(ctx, city).Return(entity.Point{}, repository.ErrNotFound)
			},
			wantErr: service.ErrCityNotFound,
		},
		{
			name: "cannot create point",
			mockBehavior: func(p *repomocks.MockPoint, m *metricmocks.MockCounter) {
				p.EXPECT().Create(ctx, city).Return(entity.Point{}, arbitraryErr)
			},
			wantErr: service.ErrCannotCreatePoint,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockPointRepo := repomocks.NewMockPoint(ctrl)
			mockProductRepo := repomocks.NewMockProduct(ctrl)
			mockReceptionRepo := repomocks.NewMockReception(ctrl)
			mockPointCounter := metricmocks.NewMockCounter(ctrl)

			tc.mockBehavior(mockPointRepo, mockPointCounter)

			s := service.NewPointService(mockPointRepo, mockProductRepo, mockReceptionRepo, mockPointCounter)

			got, err := s.Create(ctx, city)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPointService_GetAll(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		arbitraryErr = errors.New("arbitrary error")
		ctx          = context.Background()
	)

	points := []entity.Point{
		{
			ID:        uuid.New(),
			City:      "Москва",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			City:      "Санкт-Петербург",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			City:      "Казань",
			CreatedAt: time.Now(),
		},
	}

	type MockBehavior func(p *repomocks.MockPoint)

	for _, tc := range []struct {
		name         string
		mockBehavior MockBehavior
		want         []entity.Point
		wantErr      error
	}{
		{
			name: "success",
			mockBehavior: func(p *repomocks.MockPoint) {
				p.EXPECT().GetAll(ctx).Return(points, nil)
			},
			want: points,
		},
		{
			name: "cannot get all points",
			mockBehavior: func(p *repomocks.MockPoint) {
				p.EXPECT().GetAll(ctx).Return([]entity.Point{}, arbitraryErr)
			},
			want:    []entity.Point{},
			wantErr: service.ErrCannotGetPoints,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockPointRepo := repomocks.NewMockPoint(ctrl)
			mockProductRepo := repomocks.NewMockProduct(ctrl)
			mockReceptionRepo := repomocks.NewMockReception(ctrl)
			mockPointCounter := metricmocks.NewMockCounter(ctrl)

			tc.mockBehavior(mockPointRepo)

			s := service.NewPointService(mockPointRepo, mockProductRepo, mockReceptionRepo, mockPointCounter)

			got, err := s.GetAll(ctx)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPointService_GetExtended(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		arbitraryErr = errors.New("arbitrary error")
		ctx          = context.Background()
		start        = lo.Must(time.Parse(time.RFC3339, "2025-03-15T14:00:00Z"))
		end          = lo.Must(time.Parse(time.RFC3339, "2025-03-17T19:49:00Z"))
		page         = service.DefaultPage
		offset       = (service.DefaultPage - 1) * service.DefaultLimit
		limit        = service.DefaultLimit
	)

	output := []dto.PointOutput{
		{
			Point: dto.Point{
				ID:        uuid.MustParse("2ddd8d38-c1cc-4c46-acc1-f10e97c57cdd"),
				CreatedAt: lo.Must(time.Parse(time.RFC3339, "2025-04-13T08:19:45.739745Z")),
				City:      "Санкт-Петербург",
			},
			Receptions: []dto.ReceptionResult{},
		},
		{
			Point: dto.Point{
				ID:        uuid.MustParse("630aca75-ae74-4298-bd85-6629838a655a"),
				CreatedAt: lo.Must(time.Parse(time.RFC3339, "2025-04-13T08:21:32.236032Z")),
				City:      "Казань",
			},
			Receptions: []dto.ReceptionResult{
				{
					Reception: dto.Reception{
						ID:        uuid.MustParse("489eb0e1-9f6a-4ced-8e9c-2c0551a47c53"),
						PointID:   uuid.MustParse("630aca75-ae74-4298-bd85-6629838a655a"),
						CreatedAt: lo.Must(time.Parse(time.RFC3339, "2025-04-13T08:31:32.471022Z")),
						Status:    entity.ReceptionStatusClosed,
					},
					Products: []dto.Product{
						{
							ID:          uuid.MustParse("466b0c7e-af3c-478b-a533-a2621834383a"),
							ReceptionID: uuid.MustParse("489eb0e1-9f6a-4ced-8e9c-2c0551a47c53"),
							CreatedAt:   lo.Must(time.Parse(time.RFC3339, "2025-04-13T08:31:50.477739Z")),
							Type:        entity.ProductTypeShoes,
						},
						{
							ID:          uuid.MustParse("37d4f863-fee9-4629-9c51-5f1d4c23e412"),
							ReceptionID: uuid.MustParse("489eb0e1-9f6a-4ced-8e9c-2c0551a47c53"),
							CreatedAt:   lo.Must(time.Parse(time.RFC3339, "2025-04-13T08:32:06.880357Z")),
							Type:        entity.ProductTypeElectronics,
						},
						{
							ID:          uuid.MustParse("d3100895-2de8-42a2-a239-e6bd0c25eecc"),
							ReceptionID: uuid.MustParse("489eb0e1-9f6a-4ced-8e9c-2c0551a47c53"),
							CreatedAt:   lo.Must(time.Parse(time.RFC3339, "2025-04-13T08:32:13.050501Z")),
							Type:        entity.ProductTypeClothes,
						},
					},
				},
			},
		},
	}

	type args struct {
		start *time.Time
		end   *time.Time
	}
	type MockBehavior func(p *repomocks.MockPoint)

	for _, tc := range []struct {
		name         string
		args         args
		mockBehavior MockBehavior
		want         []dto.PointOutput
		wantErr      error
	}{
		{
			name: "success",
			args: args{
				start: &start,
				end:   &end,
			},
			mockBehavior: func(p *repomocks.MockPoint) {
				p.EXPECT().GetExtended(ctx, &start, &end, offset, limit).Return(output, nil)
			},
			want: output,
		},
		{
			name: "success without filters",
			args: args{
				start: nil,
				end:   nil,
			},
			mockBehavior: func(p *repomocks.MockPoint) {
				p.EXPECT().GetExtended(ctx, nil, nil, offset, limit).Return(output, nil)
			},
			want: output,
		},
		{
			name: "cannot get points",
			args: args{
				start: &start,
				end:   &end,
			},
			mockBehavior: func(p *repomocks.MockPoint) {
				p.EXPECT().GetExtended(ctx, &start, &end, offset, limit).Return([]dto.PointOutput{}, arbitraryErr)
			},
			want:    []dto.PointOutput{},
			wantErr: service.ErrCannotGetPoints,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockPointRepo := repomocks.NewMockPoint(ctrl)
			mockProductRepo := repomocks.NewMockProduct(ctrl)
			mockReceptionRepo := repomocks.NewMockReception(ctrl)
			mockPointCounter := metricmocks.NewMockCounter(ctrl)

			tc.mockBehavior(mockPointRepo)

			s := service.NewPointService(mockPointRepo, mockProductRepo, mockReceptionRepo, mockPointCounter)

			got, err := s.GetExtended(ctx, tc.args.start, tc.args.end, &page, &limit)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPointService_CloseLastReception(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		arbitraryErr = errors.New("arbitrary error")
		ctx          = context.Background()
		pointID      = uuid.New()
		receptionID  = uuid.New()
		timestamp    = time.Now().Add(-time.Hour)
	)

	reception := entity.Reception{
		ID:        receptionID,
		PointID:   pointID,
		CreatedAt: timestamp,
		Status:    entity.ReceptionStatusClosed,
	}

	type MockBehavior func(r *repomocks.MockReception)

	for _, tc := range []struct {
		name         string
		mockBehavior MockBehavior
		want         entity.Reception
		wantErr      error
	}{
		{
			name: "success",
			mockBehavior: func(r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(receptionID, nil)
				r.EXPECT().Close(ctx, receptionID).Return(reception, nil)
			},
			want: reception,
		},
		{
			name: "active reception not found",
			mockBehavior: func(r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(uuid.Nil, repository.ErrNotFound)
			},
			wantErr: service.ErrActiveReceptionNotFound,
		},
		{
			name: "cannot find reception",
			mockBehavior: func(r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(uuid.Nil, arbitraryErr)
			},
			wantErr: service.ErrCannotCloseReception,
		},
		{
			name: "cannot close reception",
			mockBehavior: func(r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(receptionID, nil)
				r.EXPECT().Close(ctx, receptionID).Return(entity.Reception{}, arbitraryErr)
			},
			wantErr: service.ErrCannotCloseReception,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockPointRepo := repomocks.NewMockPoint(ctrl)
			mockProductRepo := repomocks.NewMockProduct(ctrl)
			mockReceptionRepo := repomocks.NewMockReception(ctrl)
			mockPointCounter := metricmocks.NewMockCounter(ctrl)

			tc.mockBehavior(mockReceptionRepo)

			s := service.NewPointService(mockPointRepo, mockProductRepo, mockReceptionRepo, mockPointCounter)

			got, err := s.CloseLastReception(ctx, pointID)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPointService_DeleteLastProduct(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		arbitraryErr = errors.New("arbitrary error")
		ctx          = context.Background()
		pointID      = uuid.New()
		receptionID  = uuid.New()
		productID    = uuid.New()
	)

	type MockBehavior func(p *repomocks.MockProduct, r *repomocks.MockReception)

	for _, tc := range []struct {
		name         string
		mockBehavior MockBehavior
		wantErr      error
	}{
		{
			name: "success",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(receptionID, nil)
				p.EXPECT().GetLatestID(ctx, receptionID).Return(productID, nil)
				p.EXPECT().DeleteByID(ctx, productID).Return(nil)
			},
		},
		{
			name: "active reception not found",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(uuid.Nil, repository.ErrNotFound)
			},
			wantErr: service.ErrActiveReceptionNotFound,
		},
		{
			name: "cannot get active reception",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(uuid.Nil, arbitraryErr)
			},
			wantErr: service.ErrCannotDeleteLastProduct,
		},
		{
			name: "product not found",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(receptionID, nil)
				p.EXPECT().GetLatestID(ctx, receptionID).Return(uuid.Nil, repository.ErrNotFound)
			},
			wantErr: service.ErrProductNotFound,
		},
		{
			name: "cannot get last product",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(receptionID, nil)
				p.EXPECT().GetLatestID(ctx, receptionID).Return(uuid.Nil, arbitraryErr)
			},
			wantErr: service.ErrCannotDeleteLastProduct,
		},
		{
			name: "no rows deleted",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(receptionID, nil)
				p.EXPECT().GetLatestID(ctx, receptionID).Return(productID, nil)
				p.EXPECT().DeleteByID(ctx, productID).Return(repository.ErrNoRowsDeleted)
			},
			wantErr: service.ErrProductAlreadyDeleted,
		},
		{
			name: "cannot delete last product",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(receptionID, nil)
				p.EXPECT().GetLatestID(ctx, receptionID).Return(productID, nil)
				p.EXPECT().DeleteByID(ctx, productID).Return(arbitraryErr)
			},
			wantErr: service.ErrCannotDeleteLastProduct,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockPointRepo := repomocks.NewMockPoint(ctrl)
			mockProductRepo := repomocks.NewMockProduct(ctrl)
			mockReceptionRepo := repomocks.NewMockReception(ctrl)
			mockPointCounter := metricmocks.NewMockCounter(ctrl)

			tc.mockBehavior(mockProductRepo, mockReceptionRepo)

			s := service.NewPointService(mockPointRepo, mockProductRepo, mockReceptionRepo, mockPointCounter)

			err := s.DeleteLastProduct(ctx, pointID)

			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
