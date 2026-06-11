package image

// Configurações do preview — ajuste os valores abaixo para alterar
// resolução, compressão e aparência da marca d'água.

var (
	// PreviewMaxWidth define a largura máxima em pixels (mantém proporção).
	PreviewMaxWidth = 480

	// PreviewJPEGQuality controla a compressão JPEG (1–100; menor = pior qualidade, arquivo menor).
	PreviewJPEGQuality = 25

	// PreviewWatermarkText é o texto exibido na marca d'água.
	PreviewWatermarkText = "GO GALLERY"

	// WatermarkFontSizeRatio divide a largura da imagem para calcular o tamanho da fonte.
	// Valor maior = fonte menor (ex: 8 → largura/8).
	WatermarkFontSizeRatio = 8

	// WatermarkAlpha define a opacidade do texto (0–255; menor = mais transparente).
	WatermarkAlpha = 190

	// WatermarkGapX espaçamento horizontal extra entre repetições (pixels).
	WatermarkGapX = 30

	// WatermarkGapY espaçamento vertical extra entre repetições (pixels).
	WatermarkGapY = 50

	// WatermarkAngle ângulo de inclinação em graus (negativo = diagonal descendente).
	WatermarkAngle = -30
)
