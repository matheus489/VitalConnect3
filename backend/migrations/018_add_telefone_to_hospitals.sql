-- Migration: 018_add_telefone_to_hospitals
-- Description: Add telefone column for hospital contact information
-- Created: 2026-01-17

-- UP
-- Adicionar campo de telefone para contato do hospital (recepcao/plantao)
ALTER TABLE hospitals ADD COLUMN IF NOT EXISTS telefone VARCHAR(20);

-- Comentario descritivo
COMMENT ON COLUMN hospitals.telefone IS 'Telefone de contato do hospital - formato brasileiro: (XX) XXXX-XXXX ou (XX) XXXXX-XXXX';

-- DOWN (for rollback)
-- ALTER TABLE hospitals DROP COLUMN IF EXISTS telefone;
