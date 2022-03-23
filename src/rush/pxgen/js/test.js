import { genPx } from './env.js';

// await main_bak();
// while (1) {
var pxArgs;
try {
  pxArgs = JSON.parse(Deno.args[0]);
} catch(e) {
  pxArgs = {
    AppId: 'PXAJDckzHD',
    JsSrc: 'https://www.hibbett.com/on/demandware.store/Sites-Hibbett-US-Site/default/IZ-Client/',
    Origin: 'https://www.hibbett.com',
    UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36",
  };
  // pxArgs = {
  //   Cookie: '__cfduid=dc0f48271b210111f06800ce1f68dc76d1591143053; dwac_269962172965e441be86869251=qgfCo0WXImDusYTAdrZRPd41GSn-G9ThURU%3D|dw-only|||USD|false|America%2FChicago|true; cqcid=bcKNVmHB2C6KaJ6BEVyFkIBYiD; sid=qgfCo0WXImDusYTAdrZRPd41GSn-G9ThURU; dwanonymous_9bc4215d5ba3521ff1e6f9e5299a7444=bcKNVmHB2C6KaJ6BEVyFkIBYiD; __cq_dnt=0; dw_dnt=0; dwsid=fBDlTukaEqdnZtQE_DFELQgiPsRSluQm2MJ7N6KdOGkgNv7JgpDf5JoI2GcoAp5DsodkPjv5jFitfSUzVvTUSQ==; ku1-sid=2ZAKBZmFeHN6JQE6saUvY; ku1-vid=04f16dd1-9775-0885-92f9-57a818e1cba5; dw=1; dw_cookies_accepted=1; BVImplpixel_pie=18471_3_0; __olapicU=1591143062613; tfc-l=%7B%22a%22%3A%7B%22v%22%3A%22cd731956-cf8d-4399-9043-582c52f2ff31%22%2C%22e%22%3A1591229462%7D%2C%22u%22%3A%7B%22v%22%3A%22V6%7Cunk_20a10d2e-8b26-4190-ad11-0a4f5a0e332e!1654042261%7Ctjsc997r0r4bntnem5elaqfp3d!1654042261%22%2C%22e%22%3A1654042262%7D%2C%22s%22%3A%7B%22v%22%3A%22session.params%3D%257C1654042261%22%2C%22e%22%3A1654042262%7D%2C%22k%22%3A%7B%22v%22%3A%22tjsc997r0r4bntnem5elaqfp3d%22%2C%22e%22%3A1654042261%7D%7D; tfc-s=%7B%22v%22%3A%22tfc-fitrec-product%3D1%22%7D; liveagent_oref=; rdf-uuid=0367a651-4616-4c04-a43c-70774d4eb30f_1591083294341; rdf-count=1; liveagent_sid=96f1c8e6-8ac1-4ecf-882c-9e12003dcef7; liveagent_vc=2; liveagent_ptid=96f1c8e6-8ac1-4ecf-882c-9e12003dcef7',
  //   AppId: 'PXAJDckzHD',
  //   Vid: '',
  //   Uuid: '24527400-a533-11ea-b109-51a33c061e2c',
  //   Host: 'https://collector-pxajdckzhd.perimeterx.net',
  //   Proxy: 'http://127.0.0.1:8888',
  //   JsSrc: 'https://www.hibbett.com/on/demandware.store/Sites-Hibbett-US-Site/default/IZ-Client/',
  //   CapJsSrc: 'https://www.hibbett.com/on/demandware.store/Sites-Hibbett-US-Site/default/Product-Show/IZ/ajdckzhd/captcha?a=c&u=88071440-a530-11ea-835a-c32c2a7d54cf&v=&m=0',
  // }
}
console.log(await genPx());
//   if (DEBUG) await delay(1000);
// }
// if (!DEBUG) {
Deno.exit(0);
// }