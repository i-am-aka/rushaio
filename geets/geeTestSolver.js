// import { createRequire } from "https://deno.land/std@0.79.0/node/module.ts";
// const require = createRequire(import.meta.url);

// const path = require('path');
// const fs = require('fs');
// import * as path from 'https://deno.land/std@0.78.0/path/mod.ts';
// import * as fs from 'https://deno.land/std@0.78.0/node/fs.ts';
// import jsdom from 'http://127.0.0.1:8080/jsdom.js';
import jsdom from 'https://rushjspm.s3.amazonaws.com/jsdom.js';
const { JSDOM } = jsdom;
import { solvePuzzle } from "./findPuzzleCoord.js";
import { genMouseScale } from "./genMouseScale.js";
// virtualConsole.sendTo(console);
// var readline = require("readline");
// const URL = require("url");

function randomInRange(min, max) {
  return Math.random() * (max - min) + min;
}

function genRandom(maxVal, minVal, maxLenDeci, minLenDeci) {
  // adds decimal of certain len to emulate how browsers return values
  let a = Math.floor(Math.random() * (maxVal - minVal) + minVal);
  let len = Math.floor(Math.random() * (maxLenDeci - minLenDeci) + minLenDeci);
  let randVal = a.toString() + ".";
  for (i = 0; i < len; i++) {
    randVal += Math.floor(Math.random() * 9.9).toString();
  }
  return parseFloat(randVal);
}

function sleep(milliseconds, vmContext) {
  // dom.window.spoofedTime += milliseconds
  const date = Date.now();
  let currentDate = null;
  do {
    currentDate = Date.now();
  } while (currentDate - date < milliseconds);
}

const deepEqual = function (x, y) {
  if (x === y) {
    return true;
  } else if (
    typeof x == "object" &&
    x != null &&
    typeof y == "object" &&
    y != null
  ) {
    if (Object.keys(x).length != Object.keys(y).length) {
      return false;
    }
    for (var prop in x) {
      if (arguments[2] == "frames") {
      }
      if (prop.startsWith("_globalObject")) {
        continue;
      }
      if (y.hasOwnProperty(prop)) {
        if (!deepEqual(x[prop], y[prop])) return false;
      } else {
        return false;
      }
    }

    return true;
  } else {
    return false;
  }
};

export async function solveGeetest(url, xhrProxy) {
  // return new Promise(async (gennedGeetest, fatalErr) => {
  let closeDom = ()=>{};
  let success, fail;
  let successPromise = new Promise((resolve, reject) => {
    success = resolve;
    fail = reject;
  });
  successPromise = Promise.race([successPromise, delay(20e3)]);
  let dom;
  try {
    const startTime = Date.now();
    let score = undefined;
    class CustomResourceLoader extends jsdom.ResourceLoader {
      // add monitor blocker
      // add cached files to this like the slider/main script
      fetch(url, options) {
        // console.log('FETCH', url);
        const blackList = [
          "https://www.googletagmanager.com",
          "https://static.captcha-delivery.com/captcha/assets",
          "/check",
          "https://fonts.googleapis.com",
          "js.datadome.co",
          ".css",
        ];
        if (blackList.some((blackPhrase) => url.includes(blackPhrase)))
          return Promise.resolve((new TextEncoder()).encode(""));

        // Monitor Detection
        if (url.includes("monitor.geetest.com")) {
          const purl = new URL(url);
          const params = new URLSearchParams(purl.search);
          fail({
            error: "MONITOR TRIGGERED",
            url: url,
            options: options,
            response: params,
          });
          return Promise.resolve((new TextEncoder()).encode(""));
        }

        // // Fullpage Script
        // if (
        //   url.includes(
        //     "https://static.geetest.com/static/js/fullpage.9.0.2.js"
        //   )
        // ) {
        //   return new Promise((r, e) => {
        //     fs.readFile("./fullpage.9.0.2.js",
        //       "utf8",
        //       (err, data) => {
        //         if (err) {
        //           e({ error: err });
        //           return;
        //         }
        //         r(data);
        //       }
        //     );
        //   });
        // }

        // // Slide script
        // if (
        //   url.includes("https://static.geetest.com/static/js/slide.7.7.5.js")
        // ) {
        //   return new Promise((r, e) => {
        //     fs.readFile(
        //       "./slide.7.7.5.js",
        //       "utf8",
        //       (err, data) => {
        //         if (err) {
        //           e({ error: err });
        //           return;
        //         }
        //         r(data);
        //       }
        //     );
        //   });
        // }

        // if (url.includes("https://api-na.geetest.com/gettype.php")) {
        //   const purl = new URL(url);
        //   // console.log(purl)
        //   const params = new URLSearchParams(purl.search);
        //   const ret =  (new TextEncoder()).encode(
        //       `${params.get(
        //         "callback"
        //       )}({"status": "success", "data": {"aspect_radio": {"click": 128, "voice": 128, "slide": 103, "pencil": 128, "beeline": 50}, "pencil": "/static/js/pencil.1.0.3.js", "fullpage": "/static/js/fullpage.9.0.2.js", "geetest": "/static/js/geetest.6.0.9.js", "click": "/static/js/click.2.9.4.js", "static_servers": ["static.geetest.com/", "dn-staticdown.qbox.me/"], "beeline": "/static/js/beeline.1.0.1.js", "slide": "/static/js/slide.7.7.5.js", "voice": "/static/js/voice.1.2.0.js", "type": "fullpage", "maze": "/static/js/maze.1.0.1.js"}})`
        //     );
        //   const p = Promise.resolve(
        //     // ret
        //     `${params.get(
        //         "callback"
        //       )}({"status": "success", "data": {"aspect_radio": {"click": 128, "voice": 128, "slide": 103, "pencil": 128, "beeline": 50}, "pencil": "/static/js/pencil.1.0.3.js", "fullpage": "/static/js/fullpage.9.0.2.js", "geetest": "/static/js/geetest.6.0.9.js", "click": "/static/js/click.2.9.4.js", "static_servers": ["static.geetest.com/", "dn-staticdown.qbox.me/"], "beeline": "/static/js/beeline.1.0.1.js", "slide": "/static/js/slide.7.7.5.js", "voice": "/static/js/voice.1.2.0.js", "type": "fullpage", "maze": "/static/js/maze.1.0.1.js"}})`
        //   );
        //   // console.log(`${params.get(
        //         "callback"
        //       )}({"status": "success", "data": {"aspect_radio": {"click": 128, "voice": 128, "slide": 103, "pencil": 128, "beeline": 50}, "pencil": "/static/js/pencil.1.0.3.js", "fullpage": "/static/js/fullpage.9.0.2.js", "geetest": "/static/js/geetest.6.0.9.js", "click": "/static/js/click.2.9.4.js", "static_servers": ["static.geetest.com/", "dn-staticdown.qbox.me/"], "beeline": "/static/js/beeline.1.0.1.js", "slide": "/static/js/slide.7.7.5.js", "voice": "/static/js/voice.1.2.0.js", "type": "fullpage", "maze": "/static/js/maze.1.0.1.js"}})`)
        //   p.href = url;
        //   p.response = new Proxy({
        //     headers: {},
        //     statusCode: 200,
        //   }, {
        //     get: function() {
        //       // console.log(...arguments);
        //       return Reflect.get(...arguments);
        //     }
        //   })
        //   // p.response = {

        //   // }
        //   // p.href = url;
        //   // p.response = resp;
        //   // var p = fetch(url).then(resp => {
        //   //   // console.log({resp})
        //   //   p.href = resp.url;
        //   //   return resp.arrayBuffer()
        //   // });
        //   return p;
        // }

        if (url.includes("https://api-na.geetest.com/ajax.php")) {
          const purl = new URL(url);
          const params = new URLSearchParams(purl.search);
          return new Promise((resolve, err) => {
            super.fetch(url, options).then((body) => {
              // body = (new TextDecoder()).decode(body);
              // console.log(body);
              const resJson = JSON.parse(
                body
                  .toString()
                  .replace(params.get("callback"), "")
                  .slice(1, -1)
              );
              if (resJson["message"] == "fail") {
                fail({
                  error: "SLIDE_MESSAGE_FAIL",
                  url: url,
                  options: options,
                  response: resJson,
                });
              } else if (resJson["status"] == "error") {
                fail({
                  error: "GEETEST_API_FAIL",
                  url: url,
                  options: options,
                  response: resJson,
                });
              }
              score = resJson["score"];
              resolve(body);
            }).catch(fail);
          });
        }

        if (url.includes("https://api-na.geetest.com/")) {
          const purl = new URL(url);
          const params = new URLSearchParams(purl.search);
          return new Promise((resolve, err) => {
            super.fetch(url, options).then((body) => {
              // console.log(body
                  // .toString()
                  // .replace(params.get("callback"), "")
                  // .slice(0, -1))
              // console.log(typeof body)
              // console.log(body);
              const resJson = JSON.parse(
                body
                  .toString()
                  .replace(params.get("callback"), "")
                  .slice(1, -1)
              );
              if (resJson["status"] == "error") {
                err({
                  error: "GEETEST_API_FAIL",
                  url: url,
                  options: options,
                  response: resJson,
                });
              }
              resolve(body);
            }).catch(fail);
          });
        }
        return super.fetch(url, options);
      }
    }

    const resourceLoader = new CustomResourceLoader({
      proxy: "http://127.0.0.1:8889",
      strictSSL: false,
      userAgent:
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.66 Safari/537.36",
    });
    const virtualConsole = new jsdom.VirtualConsole();
    virtualConsole.on("error", (e) => {
      // fail({ error: e });
    });
    virtualConsole.sendTo(console);
    dom = await JSDOM.fromURL(url, {
      includeNodeLocations: false, // enable for scanning scripts, disable for performance
      storageQuota: 10000000,
      runScripts: "dangerously",
      resources: resourceLoader,
      pretendToBeVisual: true,
      virtualConsole,
      xhrProxy,
    })
    closeDom = function () {
      try {
        dom.window.close();
      } catch {}
    };
    // const vmContext = dom.getInternalVMContext();
    const rects = JSON.parse(
      `{"captcha":{"x":0,"y":0,"width":1536,"height":750.75,"top":0,"right":1536,"bottom":750.75,"left":0},"captcha__header":{"x":0,"y":20,"width":1536,"height":64.55000305175781,"top":20,"right":1536,"bottom":84.55000305175781,"left":0},"captcha__header__container":{"x":368,"y":20,"width":800,"height":64.55000305175781,"top":20,"right":1168,"bottom":84.55000305175781,"left":368},"captcha__header__logo-wrapper":{"x":668,"y":30,"width":200,"height":44.54999923706055,"top":30,"right":868,"bottom":74.54999923706055,"left":668},"captcha__header__logo":{"x":668,"y":30,"width":200,"height":40.54999923706055,"top":30,"right":868,"bottom":70.54999923706055,"left":668},"captcha__human":{"x":0,"y":84.55000305175781,"width":1536,"height":375,"top":84.55000305175781,"right":1536,"bottom":459.5500030517578,"left":0},"captcha__human__container":{"x":368,"y":169.5500030517578,"width":800,"height":205,"top":169.5500030517578,"right":1168,"bottom":374.5500030517578,"left":368},"captcha__human__title":{"x":378,"y":179.5500030517578,"width":780,"height":24,"top":179.5500030517578,"right":1158,"bottom":203.5500030517578,"left":378},"captcha__human__submit-description":{"x":378,"y":203.5500030517578,"width":780,"height":26.80000114440918,"top":203.5500030517578,"right":1158,"bottom":230.350004196167,"left":378},"captcha__human__captcha-container":{"x":378,"y":262.3500061035156,"width":780,"height":44,"top":262.3500061035156,"right":1158,"bottom":306.3500061035156,"left":378},"geetest_holder geetest_wind geetest_radar_click_ready":{"x":618,"y":262.3500061035156,"width":300,"height":44,"top":262.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_form":{"x":618,"y":262.3500061035156,"width":300,"height":0,"top":262.3500061035156,"right":918,"bottom":262.3500061035156,"left":618},"geetest_btn":{"x":618,"y":262.3500061035156,"width":300,"height":44,"top":262.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_radar_btn":{"x":618,"y":262.3500061035156,"width":300,"height":44,"top":262.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_radar":{"x":625,"y":269.3500061035156,"width":30,"height":30,"top":269.3500061035156,"right":655,"bottom":299.3500061035156,"left":625},"geetest_ring":{"x":634,"y":278.3500061035156,"width":12,"height":12,"top":278.3500061035156,"right":646,"bottom":290.3500061035156,"left":634},"geetest_small":{"x":681.0750122070312,"y":398.5375061035156,"width":111.63750457763672,"height":22.825000762939453,"top":398.5375061035156,"right":792.712516784668,"bottom":421.3625068664551,"left":681.0750122070312},"geetest_sector":{"x":620.1819458007812,"y":264.5319519042969,"width":39.63613510131836,"height":39.636085510253906,"top":264.5319519042969,"right":659.8180809020996,"bottom":304.1680374145508,"left":620.1819458007812},"geetest_cross":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_h":{"x":634,"y":277.6300048828125,"width":12,"height":0.7199951410293579,"top":277.6300048828125,"right":646,"bottom":278.35000002384186,"left":634},"geetest_v":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_dot":{"x":634,"y":278.3500061035156,"width":12,"height":12,"top":278.3500061035156,"right":646,"bottom":290.3500061035156,"left":634},"geetest_scan":{"x":634,"y":278.3500061035156,"width":12,"height":12,"top":278.3500061035156,"right":646,"bottom":290.3500061035156,"left":634},"geetest_status":{"x":640,"y":284.3500061035156,"width":0,"height":0,"top":284.3500061035156,"right":640,"bottom":284.3500061035156,"left":640},"geetest_bg":{"x":640,"y":284.3500061035156,"width":0,"height":0,"top":284.3500061035156,"right":640,"bottom":284.3500061035156,"left":640},"geetest_hook":{"x":640,"y":284.3500061035156,"width":0,"height":0,"top":284.3500061035156,"right":640,"bottom":284.3500061035156,"left":640},"geetest_ie_radar":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_radar_tip":{"x":619,"y":263.3500061035156,"width":298,"height":42,"top":263.3500061035156,"right":917,"bottom":305.3500061035156,"left":619},"geetest_radar_tip_content":{"x":665,"y":276.1499938964844,"width":68.48750305175781,"height":16,"top":276.1499938964844,"right":733.4875030517578,"bottom":292.1499938964844,"left":665},"geetest_reset_tip_content":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_radar_error_code":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_logo":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_other_offline geetest_offline":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_ghost_success":{"x":918,"y":262.3500061035156,"width":0,"height":44,"top":262.3500061035156,"right":918,"bottom":306.3500061035156,"left":918},"geetest_success_btn":{"x":758,"y":262.3500061035156,"width":160,"height":44,"top":262.3500061035156,"right":918,"bottom":306.3500061035156,"left":758},"geetest_success_box":{"x":766,"y":272.3500061035156,"width":24,"height":24,"top":272.3500061035156,"right":790,"bottom":296.3500061035156,"left":766},"geetest_success_show":{"x":766,"y":272.3500061035156,"width":24,"height":24,"top":272.3500061035156,"right":790,"bottom":296.3500061035156,"left":766},"geetest_success_pie":{"x":772.9285888671875,"y":273.4743347167969,"width":21.0185546875,"height":26.822803497314453,"top":273.4743347167969,"right":793.9471435546875,"bottom":300.2971382141113,"left":772.9285888671875},"geetest_success_filter":{"x":762.0529174804688,"y":268.40289306640625,"width":21.0185546875,"height":26.822803497314453,"top":268.40289306640625,"right":783.0714721679688,"bottom":295.2256965637207,"left":762.0529174804688},"geetest_success_mask":{"x":772.9285888671875,"y":273.4743347167969,"width":21.0185546875,"height":26.822803497314453,"top":273.4743347167969,"right":793.9471435546875,"bottom":300.2971382141113,"left":772.9285888671875},"geetest_success_correct":{"x":766,"y":268.3500061035156,"width":28,"height":28,"top":268.3500061035156,"right":794,"bottom":296.3500061035156,"left":766},"geetest_success_icon":{"x":742,"y":302.3500061035156,"width":18,"height":18,"top":302.3500061035156,"right":760,"bottom":320.3500061035156,"left":742},"geetest_success_radar_tip":{"x":759,"y":263.3500061035156,"width":158,"height":42,"top":263.3500061035156,"right":917,"bottom":305.3500061035156,"left":759},"geetest_success_radar_tip_content":{"x":805,"y":276.1499938964844,"width":0,"height":16,"top":276.1499938964844,"right":805,"bottom":292.1499938964844,"left":805},"geetest_success_radar_tip_timeinfo":{"x":815,"y":277.75,"width":0,"height":13.600000381469727,"top":277.75,"right":815,"bottom":291.3500003814697,"left":815},"geetest_success_logo":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_success_offline geetest_offline":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_slide_icon":{"x":618,"y":262.3500061035156,"width":300,"height":0,"top":262.3500061035156,"right":918,"bottom":262.3500061035156,"left":618},"geetest_wait":{"x":630,"y":279.3500061035156,"width":27,"height":9,"top":279.3500061035156,"right":657,"bottom":288.3500061035156,"left":630},"geetest_wait_dot geetest_dot_1":{"x":632,"y":281.3500061035156,"width":5,"height":5,"top":281.3500061035156,"right":637,"bottom":286.3500061035156,"left":632},"geetest_wait_dot geetest_dot_2":{"x":641,"y":281.3500061035156,"width":5,"height":5,"top":281.3500061035156,"right":646,"bottom":286.3500061035156,"left":641},"geetest_wait_dot geetest_dot_3":{"x":650,"y":281.3500061035156,"width":5,"height":5,"top":281.3500061035156,"right":655,"bottom":286.3500061035156,"left":650},"geetest_goto":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_goto_ghost":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_goto_wrap":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_goto_content":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_goto_content_tip":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_goto_cancel":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_goto_confirm":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_panel":{"x":667,"y":385.1000061035156,"width":276,"height":48.66250228881836,"top":385.1000061035156,"right":943,"bottom":433.762508392334,"left":667},"geetest_panel_ghost":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_box":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_other_offline geetest_panel_offline":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_loading":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_loading_icon":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_loading_content":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_success":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_success_box":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_success_show":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_success_pie":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_success_filter":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_success_mask":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_success_correct":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_success_icon":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_success_title":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_error":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_error_icon":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_error_title":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_error_content":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_error_code":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_error_code_text":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_footer":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_footer_logo":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_footer_copyright":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"geetest_panel_next":{"x":618,"y":306.3500061035156,"width":300,"height":0,"top":306.3500061035156,"right":918,"bottom":306.3500061035156,"left":618},"captcha__human__loader":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"captcha__robot":{"x":0,"y":459.5500183105469,"width":1536,"height":231.1999969482422,"top":459.5500183105469,"right":1536,"bottom":690.7500152587891,"left":0},"captcha__robot__container":{"x":473,"y":459.5500183105469,"width":590,"height":231.1999969482422,"top":459.5500183105469,"right":1063,"bottom":690.7500152587891,"left":473},"captcha__robot__warning":{"x":483,"y":469.5500183105469,"width":570,"height":174.40000915527344,"top":469.5500183105469,"right":1053,"bottom":643.9500274658203,"left":483},"captcha__robot__contact_support":{"x":483,"y":643.9500122070312,"width":570,"height":36.79999923706055,"top":643.9500122070312,"right":1053,"bottom":680.7500114440918,"left":483},"captcha__contact_support__submit":{"x":714.2000122070312,"y":643.9500122070312,"width":100.6624984741211,"height":16.80000114440918,"top":643.9500122070312,"right":814.8625106811523,"bottom":660.7500133514404,"left":714.2000122070312},"captcha__contact_support hidden not-displayed":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"captcha__contact_form":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"captcha_contact__radio-container":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"contact__label":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"captcha__contact__radio":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"captcha__contact__input-container":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"captcha__contact__input":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"captcha__contact__input__error not-displayed":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"captcha__contact__submit":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"captcha__footer":{"x":0,"y":690.75,"width":1536,"height":60,"top":690.75,"right":1536,"bottom":750.75,"left":0},"captcha__footer__container":{"x":368,"y":710.75,"width":800,"height":20,"top":710.75,"right":1168,"bottom":730.75,"left":368},"geetest_fullpage_click geetest_float geetest_wind geetest_slide3":{"x":666,"y":276.3500061035156,"width":0,"height":0,"top":276.3500061035156,"right":666,"bottom":276.3500061035156,"left":666},"geetest_fullpage_ghost":{"x":0,"y":0,"width":1536,"height":792,"top":0,"right":1536,"bottom":792,"left":0},"geetest_fullpage_click_wrap":{"x":666,"y":276.3500061035156,"width":0,"height":0,"top":276.3500061035156,"right":666,"bottom":276.3500061035156,"left":666},"geetest_fullpage_click_box":{"x":666,"y":148.85000610351562,"width":278,"height":285.9125061035156,"top":148.85000610351562,"right":944,"bottom":434.76251220703125,"left":666},"geetest_holder geetest_mobile geetest_ant geetest_embed":{"x":667,"y":149.85000610351562,"width":276,"height":283.9125061035156,"top":149.85000610351562,"right":943,"bottom":433.76251220703125,"left":667},"geetest_wrap":{"x":667,"y":149.85000610351562,"width":276,"height":220.375,"top":149.85000610351562,"right":943,"bottom":370.2250061035156,"left":667},"geetest_widget":{"x":675.9375,"y":158.77500915527344,"width":258.13751220703125,"height":158.85000610351562,"top":158.77500915527344,"right":934.0750122070312,"bottom":317.62501525878906,"left":675.9375},"geetest_window":{"x":675.9375,"y":158.77500915527344,"width":258.13751220703125,"height":158.85000610351562,"top":158.77500915527344,"right":934.0750122070312,"bottom":317.62501525878906,"left":675.9375},"geetest_link":{"x":675.9375,"y":158.77500915527344,"width":258.13751220703125,"height":0,"top":158.77500915527344,"right":934.0750122070312,"bottom":158.77500915527344,"left":675.9375},"geetest_canvas_img geetest_absolute":{"x":675.9375,"y":158.77500915527344,"width":258.13751220703125,"height":158.85000610351562,"top":158.77500915527344,"right":934.0750122070312,"bottom":317.62501525878906,"left":675.9375},"geetest_slicebg geetest_absolute":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_canvas_bg geetest_absolute":{"x":675.9375,"y":158.77500915527344,"width":258.13751220703125,"height":158.85000610351562,"top":158.77500915527344,"right":934.0750122070312,"bottom":317.62501525878906,"left":675.9375},"geetest_canvas_slice geetest_absolute":{"x":675.9375,"y":158.77500915527344,"width":258.13751220703125,"height":158.85000610351562,"top":158.77500915527344,"right":934.0750122070312,"bottom":317.62501525878906,"left":675.9375},"geetest_canvas_fullbg geetest_fade geetest_absolute":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_div_img geetest_absolute":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_div_bg geetest_absolute":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_div_slice geetest_absolute":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_div_fullbg geetest_fade geetest_absolute":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_refresh":{"x":675.9375,"y":158.77500915527344,"width":258.13751220703125,"height":0,"top":158.77500915527344,"right":934.0750122070312,"bottom":158.77500915527344,"left":675.9375},"geetest_refresh_tip":{"x":720.8624877929688,"y":366.5375061035156,"width":50.025001525878906,"height":22,"top":366.5375061035156,"right":770.8874893188477,"bottom":388.5375061035156,"left":720.8624877929688},"geetest_loading geetest_absolute geetest_fade":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_loading_icon":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_loading_tip":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_result":{"x":675.9375,"y":318.625,"width":258.13751220703125,"height":24,"top":318.625,"right":934.0750122070312,"bottom":342.625,"left":675.9375},"geetest_result_box":{"x":675.9375,"y":318.625,"width":258.13751220703125,"height":25.8125,"top":318.625,"right":934.0750122070312,"bottom":344.4375,"left":675.9375},"geetest_result_icon":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_result_title":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_result_content":{"x":675.9375,"y":318.625,"width":0,"height":24,"top":318.625,"right":675.9375,"bottom":342.625,"left":675.9375},"geetest_slider geetest_ready":{"x":675.9375,"y":332.5,"width":258.1125183105469,"height":37.72500228881836,"top":332.5,"right":934.0500183105469,"bottom":370.22500228881836,"left":675.9375},"geetest_slider_track":{"x":675.9375,"y":332.3625183105469,"width":239.6374969482422,"height":38,"top":332.3625183105469,"right":915.5749969482422,"bottom":370.3625183105469,"left":675.9375},"geetest_slider_tip geetest_fade":{"x":740.4625244140625,"y":332.3625183105469,"width":175.1125030517578,"height":38,"top":332.3625183105469,"right":915.5750274658203,"bottom":370.3625183105469,"left":740.4625244140625},"geetest_slider_button":{"x":669.9874877929688,"y":320.5874938964844,"width":65.5,"height":65.5,"top":320.5874938964844,"right":735.4874877929688,"bottom":386.0874938964844,"left":669.9874877929688},"geetest_close":{"x":681.0750122070312,"y":398.5375061035156,"width":19.86250114440918,"height":19.86250114440918,"top":398.5375061035156,"right":700.9375133514404,"bottom":418.4000072479248,"left":681.0750122070312},"geetest_close_tip":{"x":691.0750122070312,"y":366.5375061035156,"width":50,"height":22,"top":366.5375061035156,"right":741.0750122070312,"bottom":388.5375061035156,"left":691.0750122070312},"geetest_refresh_1":{"x":710.8624877929688,"y":398.5375061035156,"width":19.86250114440918,"height":19.86250114440918,"top":398.5375061035156,"right":730.7249889373779,"bottom":418.4000072479248,"left":710.8624877929688},"geetest_refresh_icon":{"x":710.8624877929688,"y":398.5375061035156,"width":19.86250114440918,"height":0,"top":398.5375061035156,"right":730.7249889373779,"bottom":398.5375061035156,"left":710.8624877929688},"geetest_feedback":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_feedback_icon":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_feedback_tip":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_voice":{"x":740.6500244140625,"y":398.5375061035156,"width":19.86250114440918,"height":19.86250114440918,"top":398.5375061035156,"right":760.5125255584717,"bottom":418.4000072479248,"left":740.6500244140625},"geetest_voice_tip":{"x":750.6500244140625,"y":366.5375061035156,"width":90.5,"height":22,"top":366.5375061035156,"right":841.1500244140625,"bottom":388.5375061035156,"left":750.6500244140625},"geetest_copyright":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_copyright_tip":{"x":0,"y":0,"width":0,"height":0,"top":0,"right":0,"bottom":0,"left":0},"geetest_fullpage_pointer":{"x":651,"y":276.3500061035156,"width":15,"height":0,"top":276.3500061035156,"right":666,"bottom":276.3500061035156,"left":651},"geetest_fullpage_pointer_out":{"x":651,"y":276.3500061035156,"width":16,"height":16,"top":276.3500061035156,"right":667,"bottom":292.3500061035156,"left":651},"geetest_fullpage_pointer_in":{"x":653,"y":277.3500061035156,"width":14,"height":14,"top":277.3500061035156,"right":667,"bottom":291.3500061035156,"left":653}}`
    );
    dom.window.spoofedTime = 0
    const object = dom.window.document.doctype.nextSibling;
    const property = "getBoundingClientRect";
    Object.defineProperty(object, property, {
      value: function () {
        const rVal = {
          bottom: 750.75,
          height: 750.75,
          left: 0,
          right: 1536,
          top: 0,
          width: 1536,
          x: 0,
          y: 0,
        };
        return rVal;
      },
      writable: false,
      enumerable: object.propertyIsEnumerable(property),
    });

    for (const prop in rects) {
      let observer = new dom.window.MutationObserver(function (
        mutations,
        me
      ) {
        // // console.log(mutations);
        var elem = dom.window.document.querySelector(`[class="${prop}"]`);
        if (elem) {
          Object.defineProperty(elem, "getBoundingClientRect", {
            value: function () {
              // console.trace();
              const rVal = rects[prop];
              return rVal;
            },
            writable: false,
            enumerable: elem.propertyIsEnumerable(property),
          });
          me.disconnect(); // stop observing
          return;
        }
      });
      observer.observe(dom.window.document, {
        childList: true,
        subtree: true,
      });
    }

    function randomNumber(min, max) {
      return dom.window.Math.random() * (max - min) + min;
    }

    function autoDrag(selector, distanceX) {
      const elem = dom.window.document.querySelector(selector);
      const rect = elem.getBoundingClientRect();
      const possibleStartX = { min: rect.x, max: rect.x + rect.width };
      const possibleStartY = { min: rect.y, max: rect.y + rect.height };
      const startX = randomNumber(possibleStartX.min, possibleStartX.max);
      const startY = randomNumber(possibleStartY.min, possibleStartY.max);
      const solutionRoute = genMouseScale(startX, startX + distanceX);
      var previousX = startX;
      var previousY = startY;
      const getMouseMove = function (coords) {
        return new dom.window.MouseEvent("mousemove", {
          altKey: false,
          bubbles: true,
          button: 0,
          buttons: 0,
          cancelBubble: false,
          cancelable: true,
          clientX: Math.floor(coords.x),
          clientY: Math.floor(coords.y),
          composed: true,
          ctrlKey: false,
          currentTarget: null,
          defaultPrevented: false,
          detail: 0,
          eventPhase: 0,
          fromElement: null,
          isTrusted: true,
          layerX: Math.floor(coords.x),
          layerY: Math.floor(coords.y),
          metaKey: false,
          movementX: Math.floor(coords.x - previousX),
          movementY: Math.floor(coords.y - previousY),
          offsetX: Math.floor(coords.x),
          offsetY: Math.floor(coords.y),
          pageX: Math.floor(coords.x),
          pageY: Math.floor(coords.y),
          path: [
            dom.window.document.querySelector("div.geetest_slider_button"),
            dom.window.document.querySelector(
              "div.geetest_slider.geetest_success"
            ),
            dom.window.document.querySelector("div.geetest_wrap"),
            dom.window.document.querySelector(
              "div.geetest_holder.geetest_mobile.geetest_ant.geetest_embed"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click_box"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click_wrap"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click.geetest_float.geetest_wind.geetest_slide3"
            ),
            dom.window.document.body,
            dom.window.document.querySelector("html"),
            dom.window.document,
            dom.window.window,
          ],
          relatedTarget: null,
          returnValue: true,
          screenX: Math.floor(coords.x),
          screenY: Math.floor(coords.y + 71),
          shiftKey: false,
          sourceCapabilities: { firesTouchEvents: false },
          srcElement: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          target: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          timeStamp: dom.window.performance.now(), // currentTime,
          toElement: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          type: "mousemove",
          view: dom.window.window,
          which: 0,
          x: Math.floor(coords.x),
          y: Math.floor(coords.y),
        });
      };
      const getPointerMove = function (coords, isClicking) {
        return new dom.window.MouseEvent("pointermove", {
          altKey: false,
          altitudeAngle: 1.5707963267948966,
          azimuthAngle: 0,
          bubbles: true,
          button: -1,
          buttons: isClicking ? 1 : 0,
          cancelBubble: false,
          cancelable: true,
          clientX: coords.x,
          clientY: coords.y,
          composed: true,
          ctrlKey: false,
          currentTarget: null,
          defaultPrevented: isClicking ? true : false,
          detail: 0,
          eventPhase: 0,
          fromElement: null,
          height: 1,
          isPrimary: true,
          isTrusted: true,
          layerX: Math.floor(coords.x),
          layerY: Math.floor(coords.y),
          metaKey: false,
          movementX: Math.floor(coords.x - previousX),
          movementY: Math.floor(coords.y - previousY),
          offsetX: coords.x,
          offsetY: coords.y,
          pageX: coords.x,
          pageY: coords.y,
          path: [
            dom.window.document.querySelector("div.geetest_slider_button"),
            dom.window.document.querySelector(
              "div.geetest_slider.geetest_success"
            ),
            dom.window.document.querySelector("div.geetest_wrap"),
            dom.window.document.querySelector(
              "div.geetest_holder.geetest_mobile.geetest_ant.geetest_embed"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click_box"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click_wrap"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click.geetest_float.geetest_wind.geetest_slide3"
            ),
            dom.window.document.body,
            dom.window.document.querySelector("html"),
            dom.window.document,
            dom.window.window,
          ],
          pointerId: 1,
          pointerType: "mouse",
          pressure: isClicking ? 0.5 : 0,
          relatedTarget: null,
          returnValue: isClicking ? false : true,
          screenX: coords.x,
          screenY: coords.y + 71,
          shiftKey: false,
          sourceCapabilities: null,
          srcElement: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          tangentialPressure: 0,
          target: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          tiltX: 0,
          tiltY: 0,
          timeStamp: dom.window.performance.now(), // currentTime,
          toElement: null,
          twist: 0,
          type: "pointermove",
          view: dom.window.window,
          which: 0,
          width: 1,
          x: coords.x,
          y: coords.y,
        });
      };
      const getPointerDown = function (coords) {
        return new dom.window.MouseEvent("pointerdown", {
          altKey: false,
          altitudeAngle: 1.5707963267948966,
          azimuthAngle: 0,
          bubbles: true,
          button: 0,
          buttons: 1,
          cancelBubble: false,
          cancelable: true,
          clientX: coords.x,
          clientY: coords.y,
          composed: true,
          ctrlKey: false,
          currentTarget: null,
          defaultPrevented: true,
          detail: 0,
          eventPhase: 0,
          fromElement: null,
          height: 1,
          isPrimary: true,
          isTrusted: true,
          layerX: Math.floor(coords.x),
          layerY: Math.floor(coords.y),
          metaKey: false,
          movementX: Math.floor(coords.x - previousX),
          movementY: Math.floor(coords.y - previousY),
          offsetX: coords.x,
          offsetY: coords.y,
          pageX: coords.x,
          pageY: coords.y,
          path: [
            dom.window.document.querySelector("div.geetest_slider_button"),
            dom.window.document.querySelector(
              "div.geetest_slider.geetest_success"
            ),
            dom.window.document.querySelector("div.geetest_wrap"),
            dom.window.document.querySelector(
              "div.geetest_holder.geetest_mobile.geetest_ant.geetest_embed"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click_box"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click_wrap"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click.geetest_float.geetest_wind.geetest_slide3"
            ),
            dom.window.document.body,
            dom.window.document.querySelector("html"),
            dom.window.document,
            dom.window.window,
          ],
          pointerId: 1,
          pointerType: "mouse",
          pressure: 0.5,
          relatedTarget: null,
          returnValue: false,
          screenX: coords.x,
          screenY: coords.y + 71,
          shiftKey: false,
          sourceCapabilities: null,
          srcElement: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          tangentialPressure: 0,
          target: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          tiltX: 0,
          tiltY: 0,
          timeStamp: dom.window.performance.now(), // currentTime,
          toElement: null,
          twist: 0,
          type: "pointermove",
          view: dom.window.window,
          which: 1,
          width: 1,
          x: coords.x,
          y: coords.y,
        });
      };
      const getPointerUp = function (coords) {
        return new dom.window.MouseEvent("pointerup", {
          altKey: false,
          altitudeAngle: 1.5707963267948966,
          azimuthAngle: 0,
          bubbles: true,
          button: 0,
          buttons: 0,
          cancelBubble: false,
          cancelable: true,
          clientX: coords.x,
          clientY: coords.y,
          composed: true,
          ctrlKey: false,
          currentTarget: null,
          defaultPrevented: true,
          detail: 0,
          eventPhase: 0,
          fromElement: null,
          height: 1,
          isPrimary: true,
          isTrusted: true,
          layerX: Math.floor(coords.x),
          layerY: Math.floor(coords.y),
          metaKey: false,
          movementX: Math.floor(coords.x - previousX),
          movementY: Math.floor(coords.y - previousY),
          offsetX: coords.x,
          offsetY: coords.y,
          pageX: coords.x,
          pageY: coords.y,
          path: [
            dom.window.document.querySelector("div.geetest_slider_button"),
            dom.window.document.querySelector(
              "div.geetest_slider.geetest_success"
            ),
            dom.window.document.querySelector("div.geetest_wrap"),
            dom.window.document.querySelector(
              "div.geetest_holder.geetest_mobile.geetest_ant.geetest_embed"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click_box"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click_wrap"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click.geetest_float.geetest_wind.geetest_slide3"
            ),
            dom.window.document.body,
            dom.window.document.querySelector("html"),
            dom.window.document,
            dom.window.window,
          ],
          pointerId: 1,
          pointerType: "mouse",
          pressure: 0,
          relatedTarget: null,
          returnValue: false,
          screenX: coords.x,
          screenY: coords.y + 71,
          shiftKey: false,
          sourceCapabilities: null,
          srcElement: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          tangentialPressure: 0,
          target: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          tiltX: 0,
          tiltY: 0,
          timeStamp: dom.window.performance.now(), // currentTime,
          toElement: null,
          twist: 0,
          type: "pointerup",
          view: dom.window.window,
          which: 1,
          width: 1,
          x: coords.x,
          y: coords.y,
        });
      };
      const getClick = function (coords) {
        return new dom.window.MouseEvent("pointerup", {
          altKey: false,
          bubbles: true,
          button: 0,
          buttons: 0,
          cancelBubble: false,
          cancelable: true,
          clientX: Math.floor(coords.x),
          clientY: Math.floor(coords.y),
          composed: true,
          ctrlKey: false,
          currentTarget: null,
          defaultPrevented: false,
          detail: 1,
          eventPhase: 0,
          fromElement: null,
          isTrusted: true,
          layerX: Math.floor(coords.x),
          layerY: Math.floor(coords.y),
          metaKey: false,
          movementX: Math.floor(coords.x - previousX),
          movementY: Math.floor(coords.y - previousY),
          offsetX: Math.floor(coords.x),
          offsetY: Math.floor(coords.y),
          pageX: coords.x,
          pageY: coords.y,
          path: [
            dom.window.document.querySelector("div.geetest_slider_button"),
            dom.window.document.querySelector(
              "div.geetest_slider.geetest_success"
            ),
            dom.window.document.querySelector("div.geetest_wrap"),
            dom.window.document.querySelector(
              "div.geetest_holder.geetest_mobile.geetest_ant.geetest_embed"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click_box"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click_wrap"
            ),
            dom.window.document.querySelector(
              "div.geetest_fullpage_click.geetest_float.geetest_wind.geetest_slide3"
            ),
            dom.window.document.body,
            dom.window.document.querySelector("html"),
            dom.window.document,
            dom.window.window,
          ],
          relatedTarget: null,
          returnValue: true,
          screenX: Math.floor(coords.x),
          screenY: Math.floor(coords.y + 71),
          shiftKey: false,
          sourceCapabilities: { firesTouchEvents: false },
          srcElement: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          target: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          timeStamp: dom.window.performance.now(), // currentTime,
          toElement: dom.window.document.querySelector(
            "div.geetest_slider_button"
          ),
          type: "click",
          view: dom.window.window,
          which: 1,
          x: Math.floor(coords.x),
          y: Math.floor(coords.y),
        });
      };
      const dispatchSlideEvt = function (event) {
        dom.window.window.dispatchEvent(event);
        dom.window.window.document.dispatchEvent(event);
        elem.dispatchEvent(event);
      };
      let fPointMove = getPointerMove(solutionRoute[0], false);
      let fMouseMove = getMouseMove(solutionRoute[0]);
      let pointDown = getPointerDown(solutionRoute[0]);
      dispatchSlideEvt(fPointMove);
      dispatchSlideEvt(fMouseMove);
      dispatchSlideEvt(pointDown);
      solutionRoute.forEach((coords) => {
        if (
          Math.floor(coords.x) == Math.floor(previousX) &&
          Math.floor(coords.y) == Math.floor(previousY)
        ) {
          return;
        }
        sleep(coords.delay, dom.window);
        let currentEvt = getPointerMove(coords, true);
        dispatchSlideEvt(currentEvt);
        previousX = coords.x;
        previousY = coords.y;
      });
      let pointUp = getPointerUp(solutionRoute[solutionRoute.length - 1]);
      let finalClick = getClick(solutionRoute[solutionRoute.length - 1]);
      dispatchSlideEvt(pointUp);
      dispatchSlideEvt(finalClick);
    }


    function getPathTo(element) {
        if (element.id!=='')
            return 'id("'+element.id+'")';
        if (element===dom.window.document.body)
            return element.tagName;

        var ix= 0;
        var siblings= element.parentNode.childNodes;
        for (var i= 0; i<siblings.length; i++) {
            var sibling= siblings[i];
            if (sibling===element)
                return getPathTo(element.parentNode)+'/'+element.tagName+'['+(ix+1)+']';
            if (sibling.nodeType===1 && sibling.tagName===element.tagName)
                ix++;
        }
    }

    function waitForSelector(selector) {
      return new Promise((loaded) => {
        var observer = new dom.window.MutationObserver((mutations, me) => {
          // // console.trace();
          // // console.log(selector, mutations.map(m => [m.type, m.target.tagName, m.target.id, m.target.className, getPathTo(m.target)]));
          var elem = dom.window.document.querySelector(selector);
          if (elem) {
            // console.log('waitForSelector HIT')
            loaded(elem);
            me.disconnect(); // stop observing
            return;
          }
        });
        observer.observe(dom.window.document, {
          childList: true,
          subtree: true,
        });
      });
    }

    function handleLoad(elem) {
      // console.trace();
      elem.click();
      waitForSelector('[class="geetest_canvas_bg geetest_absolute"]').then(
        (elem) => {
          // console.log('handleLoad got', elem.toDataURL("image/png"))
          var observer = new dom.window.MutationObserver(function (
            mutations,
            me
          ) {
            // // console.trace();
            // // console.log(selector, mutations.map(m => [m.type, m.target.tagName, m.target.id, m.target.className, getPathTo(m.target)]));
            if (
              elem.toDataURL("image/png") !=
              dom.window.document
                .querySelector(
                  '[class="geetest_canvas_fullbg geetest_fade geetest_absolute"]'
                )
                .toDataURL("image/png")
            ) {
              // console.log('handleLoad.inner got')
              me.disconnect(); // stop observing
              solvePuzzle(
                elem.toDataURL("image/png"),
                dom.window.document
                  .querySelector(
                    '[class="geetest_canvas_fullbg geetest_fade geetest_absolute"]'
                  )
                  .toDataURL("image/png")
              ).then((rawDistance) => {
                autoDrag(
                  '[class="geetest_slider_button"]',
                  rawDistance - 7
                );
              }).catch(fail);
              return;
            }
          });
          observer.observe(elem, {
            attributes: true,
          });
        }
      ).catch(err => fail(err));
    }

    const successObserver = new dom.window.MutationObserver(
      (mutations, me) => {
        const chal = dom.window.document
          .querySelector('[name="geetest_challenge"]')
          .getAttribute("value");
        const solution = dom.window.document
          .querySelector('[name="geetest_validate"]')
          .getAttribute("value");
        const seccode = dom.window.document
          .querySelector('[name="geetest_seccode"]')
          .getAttribute("value");
        if (chal != null && solution != null && seccode != null) {
          const solveTime = Date.now() - startTime + "ms";
          me.disconnect(); // stop observing
          success({
            geetest_challenge: chal,
            geetest_seccode: seccode,
            geetest_validate: solution,
            score: score,
            solveTime: solveTime,
          });
          return;
        }
      }
    );
    waitForSelector('[class="geetest_form"]').then((elem) => {
      // console.log('got form');
      successObserver.observe(elem, {
        childList: true,
        subtree: true,
        attributes: true,
      });
    }).catch(fail);
    waitForSelector('[class="geetest_radar_tip"]').then((elem) => {
      // console.log('got radar tip');
      handleLoad(elem);
    }).catch(fail);
    return await successPromise;
  } catch(e) {
    throw e;
  } finally {
    try{
      dom && dom.window.close();
    }catch(e){
      console.warn(e);
    }
  };
}

if (import.meta.main) {

  debugger;
  var geets = await Promise.race([delay(10e3), solveGeetest(
    "https://geo.captcha-delivery.com/captcha/?initialCid=AHrlqAAAAAMABn9uMAGvAC0ArsxS9A==&hash=A55FBF4311ED6F1BF9911EB71931D5&t=fe&s=17434&cid=7NBe-0Y_TMVbaM5v85yJ7N-x-eyEvBX8B86~5p47t0yGVzS7tngdp8fdTmd~wUPH3DejcH0vkFIx99QCi-XUTFxk147XTta55OrFrjMioC"
  )]);
  console.log(geets)
}
/* var rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false,
});

rl.on('line', (input) => {
  try {
    solveGeetest('https://geo.captcha-delivery.com/captcha/?initialCid=AHrlqAAAAAMABn9uMAGvAC0ArsxS9A==&hash=A55FBF4311ED6F1BF9911EB71931D5&t=fe&s=17434&cid=7NBe-0Y_TMVbaM5v85yJ7N-x-eyEvBX8B86~5p47t0yGVzS7tngdp8fdTmd~wUPH3DejcH0vkFIx99QCi-XUTFxk147XTta55OrFrjMioC').then(genned => {
      // console.log(genned)
    })
    // console.log('started jsdom')
  } catch (error) {
    // console.log(error)
  }
}); */
function delay(ms) {
    return new Promise((resolve, reject) => {
      setTimeout(resolve, ms);
    });
  }