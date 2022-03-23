import Jimp from 'https://dev.jspm.io/jimp';
import { decode } from 'https://deno.land/std@0.76.0/encoding/base64.ts';

const imageToPixels = (b64) => {
  // console.log(b64.split(",")[1]);
  // console.log(decode(b64.split(",")[1]));
  return new Promise((resolve, reject) => {
    try{
      Jimp.read(decode(b64.split(",")[1]).buffer).then(image => {
        try{
          // if (e) {
          //   reject(e);
          // }
          // console.log({e, image})
          const pixels = [];
          for (var x = 0; x < image.bitmap.width; x++) {
            const row = [];
            for (var y = 0; y < image.bitmap.height; y++) {
              row.push(Jimp.intToRGBA(image.getPixelColor(x, y)));
            }
            pixels.push(row);
          }
          resolve(pixels);
        }catch(e) {
          reject(e);
        }
      }).catch(reject);
    }catch(e) {
      reject(e);
    }
  });
};

const findOffset = (aPixels, bPixels) => {
  return new Promise((resolve, reject) => {
    aPixels.forEach((aColumn, x) => {
      aColumn.forEach((aPixel, y) => {
        const bPixel = bPixels[x][y];
        const rDiff = bPixel.r - aPixel.r;
        const gDiff = bPixel.g - aPixel.g;
        const bDiff = bPixel.b - aPixel.b;
        const mean = (rDiff + gDiff + bDiff) / 3;
        if (mean > 70) {
          resolve(x);
        }
      });
    });
  });
};

export const solvePuzzle = (puzzleImageB64, fullImageB64) => {
  // console.log('solvePuzzle', {puzzleImageB64, fullImageB64});

  var aPixels = [];
  var bPixels = [];
  return new Promise((resolve, reject) => {
    imageToPixels(puzzleImageB64)
      .then((v) => {
        aPixels = v;
      })
      .then(() => {
        imageToPixels(fullImageB64)
          .then((v) => {
            bPixels = v;
          })
          .then(() => {
            findOffset(aPixels, bPixels).then((offset) => {
              resolve(offset);
            });
          });
      }).catch(reject);
  });
};

