-- Migration: 017_add_coordinates_to_hospitals
-- Description: Add latitude and longitude columns for geographic map feature
-- Created: 2026-01-16

-- UP
-- Adicionar campos de coordenadas geograficas para o Dashboard Geografico
ALTER TABLE hospitals ADD COLUMN IF NOT EXISTS latitude DECIMAL(10, 8);
ALTER TABLE hospitals ADD COLUMN IF NOT EXISTS longitude DECIMAL(11, 8);

-- Indice para buscas geograficas (hospitais ativos com coordenadas)
CREATE INDEX IF NOT EXISTS idx_hospitals_coordinates
    ON hospitals(latitude, longitude)
    WHERE deleted_at IS NULL AND ativo = true AND latitude IS NOT NULL AND longitude IS NOT NULL;

-- Comentarios descritivos
COMMENT ON COLUMN hospitals.latitude IS 'Latitude geografica do hospital (formato decimal, ex: -16.6868)';
COMMENT ON COLUMN hospitals.longitude IS 'Longitude geografica do hospital (formato decimal, ex: -49.2648)';

-- DOWN (for rollback)
-- DROP INDEX IF EXISTS idx_hospitals_coordinates;
-- ALTER TABLE hospitals DROP COLUMN IF EXISTS longitude;
-- ALTER TABLE hospitals DROP COLUMN IF EXISTS latitude;
