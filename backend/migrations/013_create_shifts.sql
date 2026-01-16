-- Migration: 008_create_shifts
-- Description: Create shifts table for operator scheduling and notification routing
-- Created: 2026-01-16

-- UP
CREATE TABLE IF NOT EXISTS shifts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hospital_id UUID NOT NULL REFERENCES hospitals(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performant queries
CREATE INDEX IF NOT EXISTS idx_shifts_hospital_id ON shifts(hospital_id);
CREATE INDEX IF NOT EXISTS idx_shifts_user_id ON shifts(user_id);
CREATE INDEX IF NOT EXISTS idx_shifts_day_of_week ON shifts(day_of_week);
CREATE INDEX IF NOT EXISTS idx_shifts_hospital_day ON shifts(hospital_id, day_of_week);

-- Unique constraint to prevent duplicate shifts
CREATE UNIQUE INDEX IF NOT EXISTS idx_shifts_unique_schedule
    ON shifts(hospital_id, user_id, day_of_week, start_time);

-- Comments
COMMENT ON TABLE shifts IS 'Escalas de plantao dos operadores por hospital';
COMMENT ON COLUMN shifts.hospital_id IS 'Hospital ao qual esta escala pertence';
COMMENT ON COLUMN shifts.user_id IS 'Operador escalado para este turno';
COMMENT ON COLUMN shifts.day_of_week IS 'Dia da semana (0=Domingo, 1=Segunda, ..., 6=Sabado)';
COMMENT ON COLUMN shifts.start_time IS 'Horario de inicio do turno (ex: 07:00, 19:00)';
COMMENT ON COLUMN shifts.end_time IS 'Horario de fim do turno (ex: 19:00, 07:00 para turno noturno)';

-- DOWN (for rollback)
-- DROP TABLE IF EXISTS shifts;
