import 'package:image/image.dart' as imglib;

void main() {
  imglib.Image img = imglib.Image(width: 100, height: 100);
  imglib.drawString(img, "hello", font: imglib.arial24, x: 10, y: 10, color: imglib.ColorRgb8(255, 255, 255));
}
