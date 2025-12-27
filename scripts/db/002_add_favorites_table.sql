-- Migración: Agregar tabla de favoritos
-- Fecha: 2025-12-27
-- Descripción: Tabla para almacenar los videos favoritos de los usuarios

CREATE TABLE IF NOT EXISTS favorites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, video_id)
);

-- Índices para mejorar el rendimiento
CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_video_id ON favorites(video_id);
CREATE INDEX IF NOT EXISTS idx_favorites_created_at ON favorites(created_at DESC);

-- Comentarios
COMMENT ON TABLE favorites IS 'Almacena los videos marcados como favoritos por los usuarios';
COMMENT ON COLUMN favorites.user_id IS 'ID del usuario que marcó el video como favorito';
COMMENT ON COLUMN favorites.video_id IS 'ID del video marcado como favorito';
COMMENT ON COLUMN favorites.created_at IS 'Fecha y hora en que se marcó como favorito';
