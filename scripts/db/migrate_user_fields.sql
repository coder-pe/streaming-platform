-- Copyright (c) 2024 Miguel Mamani
-- Email: miguel.coder.per@gmail.com
-- License: MIT

-- Migración para cambiar firstName/lastName a first_name/last_name

-- Verificar si las columnas antiguas existen y renombrarlas
DO $$
BEGIN
    -- Renombrar firstName a first_name si existe
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
        AND column_name = 'firstName'
    ) THEN
        ALTER TABLE users RENAME COLUMN "firstName" TO first_name;
        RAISE NOTICE 'Columna firstName renombrada a first_name';
    END IF;

    -- Renombrar lastName a last_name si existe
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
        AND column_name = 'lastName'
    ) THEN
        ALTER TABLE users RENAME COLUMN "lastName" TO last_name;
        RAISE NOTICE 'Columna lastName renombrada a last_name';
    END IF;

    -- Verificar que las columnas ahora existen con el nombre correcto
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
        AND column_name = 'first_name'
    ) THEN
        RAISE EXCEPTION 'ERROR: Columna first_name no existe después de la migración';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
        AND column_name = 'last_name'
    ) THEN
        RAISE EXCEPTION 'ERROR: Columna last_name no existe después de la migración';
    END IF;

    RAISE NOTICE 'Migración completada exitosamente';
END $$;
