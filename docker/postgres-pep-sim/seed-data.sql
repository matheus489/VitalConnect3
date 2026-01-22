-- SIDOT PEP Simulator - Seed Data
-- Realistic Brazilian patient data for demo purposes
-- Names and data are completely fictional

-- Helper function to generate valid Brazilian CPF check digits
CREATE OR REPLACE FUNCTION generate_cpf_checkdigit(base VARCHAR(9))
RETURNS VARCHAR(11) AS $$
DECLARE
    digits INT[];
    sum1 INT := 0;
    sum2 INT := 0;
    d1 INT;
    d2 INT;
    i INT;
BEGIN
    -- Parse base digits
    FOR i IN 1..9 LOOP
        digits[i] := CAST(SUBSTRING(base FROM i FOR 1) AS INT);
    END LOOP;

    -- Calculate first check digit
    FOR i IN 1..9 LOOP
        sum1 := sum1 + digits[i] * (11 - i);
    END LOOP;
    d1 := CASE WHEN (sum1 % 11) < 2 THEN 0 ELSE 11 - (sum1 % 11) END;

    -- Calculate second check digit
    digits[10] := d1;
    FOR i IN 1..10 LOOP
        sum2 := sum2 + digits[i] * (12 - i);
    END LOOP;
    d2 := CASE WHEN (sum2 % 11) < 2 THEN 0 ELSE 11 - (sum2 % 11) END;

    RETURN SUBSTRING(base FROM 1 FOR 3) || '.' ||
           SUBSTRING(base FROM 4 FOR 3) || '.' ||
           SUBSTRING(base FROM 7 FOR 3) || '-' ||
           CAST(d1 AS VARCHAR) || CAST(d2 AS VARCHAR);
END;
$$ LANGUAGE plpgsql;

-- Helper function to generate valid CNS (15 digits)
CREATE OR REPLACE FUNCTION generate_cns()
RETURNS VARCHAR(15) AS $$
BEGIN
    -- CNS starting with 7, 8, or 9 are definitive numbers
    RETURN '7' || LPAD(CAST(FLOOR(RANDOM() * 99999999999999)::BIGINT AS VARCHAR), 14, '0');
END;
$$ LANGUAGE plpgsql;

-- Insert initial seed data (realistic Brazilian names and medical data)
INSERT INTO TASY.TB_PACIENTE_OBITO (
    CD_PACIENTE, NM_PACIENTE, DT_NASCIMENTO, NR_CNS, NR_CPF,
    DT_OBITO, DS_CAUSA_MORTIS, CD_CID, CD_SETOR, NR_LEITO, NR_PRONTUARIO,
    IE_IDENTIFICACAO_DESCONHECIDA
) VALUES
-- Initial batch of realistic demo data
(
    100001,
    'Maria da Silva Santos',
    '1945-03-15',
    generate_cns(),
    generate_cpf_checkdigit('123456789'),
    CURRENT_TIMESTAMP - INTERVAL '2 hours',
    'Infarto agudo do miocardio com supradesnivelamento do segmento ST',
    'I21.0',
    'UTI CARDIOLOGICA',
    'UC-01',
    'PRO-2024-00001',
    'N'
),
(
    100002,
    'Jose Carlos Oliveira',
    '1958-07-22',
    generate_cns(),
    generate_cpf_checkdigit('234567891'),
    CURRENT_TIMESTAMP - INTERVAL '4 hours',
    'Pneumonia bacteriana nao especificada',
    'J18.9',
    'UTI GERAL',
    'UG-05',
    'PRO-2024-00002',
    'N'
),
(
    100003,
    'Ana Paula Ferreira Lima',
    '1972-11-08',
    generate_cns(),
    generate_cpf_checkdigit('345678912'),
    CURRENT_TIMESTAMP - INTERVAL '1 hour',
    'Acidente vascular cerebral hemorragico',
    'I61.9',
    'UTI NEUROLOGICA',
    'UN-03',
    'PRO-2024-00003',
    'N'
),
(
    100004,
    'Francisco Almeida Souza',
    '1938-02-28',
    generate_cns(),
    generate_cpf_checkdigit('456789123'),
    CURRENT_TIMESTAMP - INTERVAL '30 minutes',
    'Insuficiencia respiratoria aguda secundaria a DPOC',
    'J96.0',
    'EMERGENCIA',
    'EM-12',
    'PRO-2024-00004',
    'N'
),
(
    100005,
    'Paciente Nao Identificado',
    '1980-01-01', -- Estimated date
    NULL,
    NULL,
    CURRENT_TIMESTAMP - INTERVAL '5 hours',
    'Trauma cranioencefalico grave - multiplas lesoes',
    'S06.9',
    'POLITRAUMA',
    'PT-02',
    'PRO-2024-00005',
    'S'
);

-- Array of Brazilian first names for random generation
CREATE OR REPLACE FUNCTION random_first_name()
RETURNS VARCHAR AS $$
DECLARE
    names VARCHAR[] := ARRAY[
        'Antonio', 'Jose', 'Joao', 'Francisco', 'Carlos',
        'Paulo', 'Pedro', 'Lucas', 'Luiz', 'Marcos',
        'Maria', 'Ana', 'Francisca', 'Adriana', 'Juliana',
        'Marcia', 'Fernanda', 'Patricia', 'Aline', 'Sandra',
        'Roberto', 'Ricardo', 'Fernando', 'Eduardo', 'Rafael',
        'Claudia', 'Cristina', 'Lucia', 'Helena', 'Tereza'
    ];
BEGIN
    RETURN names[1 + FLOOR(RANDOM() * ARRAY_LENGTH(names, 1))::INT];
END;
$$ LANGUAGE plpgsql;

-- Array of Brazilian last names for random generation
CREATE OR REPLACE FUNCTION random_last_name()
RETURNS VARCHAR AS $$
DECLARE
    names VARCHAR[] := ARRAY[
        'Silva', 'Santos', 'Oliveira', 'Souza', 'Lima',
        'Pereira', 'Ferreira', 'Costa', 'Rodrigues', 'Almeida',
        'Nascimento', 'Carvalho', 'Gomes', 'Martins', 'Araujo',
        'Ribeiro', 'Barbosa', 'Andrade', 'Dias', 'Moreira',
        'Vieira', 'Alves', 'Monteiro', 'Cardoso', 'Mendes'
    ];
BEGIN
    RETURN names[1 + FLOOR(RANDOM() * ARRAY_LENGTH(names, 1))::INT];
END;
$$ LANGUAGE plpgsql;

-- Array of causes of death for random generation
CREATE OR REPLACE FUNCTION random_causa_mortis()
RETURNS TABLE(descricao VARCHAR, cid VARCHAR) AS $$
DECLARE
    causas VARCHAR[][] := ARRAY[
        ARRAY['Infarto agudo do miocardio', 'I21.9'],
        ARRAY['Acidente vascular cerebral isquemico', 'I63.9'],
        ARRAY['Acidente vascular cerebral hemorragico', 'I61.9'],
        ARRAY['Pneumonia bacteriana', 'J18.9'],
        ARRAY['Insuficiencia respiratoria aguda', 'J96.0'],
        ARRAY['Sepse grave', 'A41.9'],
        ARRAY['Trauma cranioencefalico', 'S06.9'],
        ARRAY['Insuficiencia cardiaca congestiva', 'I50.9'],
        ARRAY['Choque hipovolemico', 'R57.1'],
        ARRAY['Embolia pulmonar', 'I26.9'],
        ARRAY['Insuficiencia renal aguda', 'N17.9'],
        ARRAY['Hemorragia digestiva alta', 'K92.2']
    ];
    idx INT;
BEGIN
    idx := 1 + FLOOR(RANDOM() * ARRAY_LENGTH(causas, 1))::INT;
    descricao := causas[idx][1];
    cid := causas[idx][2];
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- Array of hospital sectors for random generation
CREATE OR REPLACE FUNCTION random_setor()
RETURNS TABLE(setor VARCHAR, leito_prefix VARCHAR) AS $$
DECLARE
    setores VARCHAR[][] := ARRAY[
        ARRAY['UTI GERAL', 'UG'],
        ARRAY['UTI CARDIOLOGICA', 'UC'],
        ARRAY['UTI NEUROLOGICA', 'UN'],
        ARRAY['EMERGENCIA', 'EM'],
        ARRAY['CENTRO CIRURGICO', 'CC'],
        ARRAY['POLITRAUMA', 'PT'],
        ARRAY['CLINICA MEDICA', 'CM'],
        ARRAY['CARDIOLOGIA', 'CA']
    ];
    idx INT;
BEGIN
    idx := 1 + FLOOR(RANDOM() * ARRAY_LENGTH(setores, 1))::INT;
    setor := setores[idx][1];
    leito_prefix := setores[idx][2];
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- Function to insert a random death record (called by cron job)
CREATE OR REPLACE FUNCTION insert_random_obito()
RETURNS VOID AS $$
DECLARE
    v_nome VARCHAR;
    v_nascimento DATE;
    v_causa RECORD;
    v_setor RECORD;
    v_cpf_base VARCHAR;
    v_prontuario VARCHAR;
    v_paciente_id BIGINT;
    v_is_unknown BOOLEAN;
BEGIN
    -- 10% chance of unidentified patient
    v_is_unknown := RANDOM() < 0.1;

    -- Generate patient ID
    v_paciente_id := 100000 + FLOOR(RANDOM() * 900000)::INT;

    -- Generate random birth date (ages 20-95)
    v_nascimento := CURRENT_DATE - (INTERVAL '1 year' * (20 + FLOOR(RANDOM() * 75)::INT));

    -- Generate random name
    v_nome := CASE
        WHEN v_is_unknown THEN 'Paciente Nao Identificado'
        ELSE random_first_name() || ' ' || random_last_name() || ' ' || random_last_name()
    END;

    -- Generate random CPF base (9 digits)
    v_cpf_base := LPAD(CAST(FLOOR(RANDOM() * 999999999)::BIGINT AS VARCHAR), 9, '0');

    -- Generate prontuario
    v_prontuario := 'PRO-' || TO_CHAR(CURRENT_DATE, 'YYYY') || '-' || LPAD(CAST(FLOOR(RANDOM() * 99999)::INT AS VARCHAR), 5, '0');

    -- Get random causa mortis
    SELECT * INTO v_causa FROM random_causa_mortis();

    -- Get random setor
    SELECT * INTO v_setor FROM random_setor();

    -- Insert record
    INSERT INTO TASY.TB_PACIENTE_OBITO (
        CD_PACIENTE, NM_PACIENTE, DT_NASCIMENTO, NR_CNS, NR_CPF,
        DT_OBITO, DS_CAUSA_MORTIS, CD_CID, CD_SETOR, NR_LEITO, NR_PRONTUARIO,
        IE_IDENTIFICACAO_DESCONHECIDA
    ) VALUES (
        v_paciente_id,
        v_nome,
        v_nascimento,
        CASE WHEN v_is_unknown THEN NULL ELSE generate_cns() END,
        CASE WHEN v_is_unknown THEN NULL ELSE generate_cpf_checkdigit(v_cpf_base) END,
        CURRENT_TIMESTAMP,
        v_causa.descricao,
        v_causa.cid,
        v_setor.setor,
        v_setor.leito_prefix || '-' || LPAD(CAST(FLOOR(RANDOM() * 20 + 1)::INT AS VARCHAR), 2, '0'),
        v_prontuario,
        CASE WHEN v_is_unknown THEN 'S' ELSE 'N' END
    );

    RAISE NOTICE 'Inserted new obito for: %', v_nome;
END;
$$ LANGUAGE plpgsql;

-- Verify seed data
DO $$
BEGIN
    RAISE NOTICE 'Seed data inserted. Total records: %', (SELECT COUNT(*) FROM TASY.TB_PACIENTE_OBITO);
END $$;
