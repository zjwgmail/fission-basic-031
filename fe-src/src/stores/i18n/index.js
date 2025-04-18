// 语言类型数字字母映射匹配
export const LANGUAGE_MODE = {
  "01": "cn", // 中文
  "02": "en", // 英文
  "03": "my", // 马来文
  "04": "id", // 印尼文
  "05": "hi", // 印度文
  "06": "tr", // 土耳其
  "07": "ru", // 俄语
  "08": "ar", // 阿拉伯语
};
// 游戏奖品类型
const GAME_PRICE_LANGUAGE_MODE = {
  "01": "3",
  "02": "5",
  "03": "8"
};

// 官方网页-礼包说明 internationalization 3、8
export const i18n = {
  data: {
    "en": {
      "activeRuleContent": [{
        "text": "1. This gift code is valid from 03/01/2025 to 04/08/2025 (UTC-8). Please use it as soon as possible before it expires."
      }, {
        "text": "2. Redeem the gift code for a random reward. Each code can only be redeemed once. Do not share or disclose this page or the gift code to others."
      }, {
        "text": "3. How to redeem: Open MLBB, tap your avatar in the top-left corner to enter your profile, and then find [Redeem Code] in settings (top-right corner) to redeem.",
        "imgs": [{
          url: "images/en/content_img_1.png"
        }]
      }, {
        "text"({ lang = "02", mode, code }) {
          let pathname = location.pathname;
          return `4. For questions about the event, <a class="goRulePage" target="_blank" href="${pathname}?lp=1&gpt=11&lang=${lang}&mode=${mode}">please check event rules ></a>`
        }
      }]
    },
    "my": {
      "activeRuleContent": [{
        "text": "1. Kod hadiah ini sah dari 01/03/2025 hingga 08/04/2025 (UTC-8). Sila gunakan kod ini secepat mungkin sebelum tamat tempoh."
      }, {
        "text": "2. Tebus kod hadiah untuk ganjaran rawak. Setiap kod hanya boleh ditebus sekali sahaja. Jangan kongsi atau dedahkan halaman ini atau kod hadiah kepada orang lain."
      }, {
        "text": "3. Cara untuk menebus: Buka MLBB, tekan avatar anda di sudut kiri atas untuk memasuki profil anda, dan kemudian cari [Tebus Kod] dalam tetapan (sudut kanan atas) untuk menebus.",
        "imgs": [{
          url: "images/my/content_img_1.png"
        }]
      }, {
        "text"({ lang = "02", mode, code }) {
          let pathname = location.pathname;
          return `4. Untuk sebarang pertanyaan mengenai acara, <a class="goRulePage" target="_blank" href="${pathname}?lp=1&gpt=11&lang=${lang}&mode=${mode}">sila semak peraturan acara ></a>`
        }
      }]
    },
    "id": {
      "activeRuleContent": [{
        "text": "1. Kode hadiah ini berlaku mulai 01/03/2025 - 08/04/2025 (UTC-8). Harap gunakan sebelum kedaluwarsa."
      }, {
        "text": "2. Tukarkan kode hadiah dengan hadiah acak. Setiap kode hanya bisa ditukar sekali. Jangan membagikan atau menyebarkan halaman atau kode hadiah ini kepada orang lain."
      }, {
        "text": "3. Cara menukar: Buka MLBB, ketuk avatar-mu di pojok kiri atas untuk masuk ke profilmu, dan ketuk [Kode Penukaran] di pengaturan (pojok kanan atas) untuk menukar.",
        "imgs": [{
          url: "images/id/content_img_1.png"
        }]
      }, {
        "text"({ lang = "02", mode, code }) {
          let pathname = location.pathname;
          return `4. Untuk pertanyaan tentang event, <a class="goRulePage" target="_blank" href="${pathname}?lp=1&gpt=11&lang=${lang}&mode=${mode}">harap cek peraturan event ></a>`
        }
      }]
    }
  }
}


// 获取国际化语言配置数据
export function queryInternationLang(langType = "02", mode) {
  return i18n.data[LANGUAGE_MODE[langType]];
  // return mode == 1 || mode == 5 ? i18n_mode1_5.data[LANGUAGE_MODE[langType]] : i18n.data[LANGUAGE_MODE[langType]];
}

// 处理国际化数据
export async function handlerInternationalizationTransform(configs = {}) {
  for (let key in configs) {
    let item = configs[key];
    // console.log(key, item, item.activeRuleContent);
    if (!Object.keys(item).length) {
      continue;
    }

    if (!!item?.langthTitleImg) {
      let imgUrl = await import(`@assets/${item.langthTitleImg}`);
      item.langthTitleImg = imgUrl.default;
    }
    if (!!item?.activeRuleContentTitleImg) {
      let imgUrl = await import(`@assets/${item.activeRuleContentTitleImg}`);
      item.activeRuleContentTitleImg = imgUrl.default;
    }
    if (!!item?.activeRuleWinningInfo) {
      let imgUrl = await import(`@assets/${item.activeRuleWinningInfo}`);
      item.activeRuleWinningInfo = imgUrl.default;
    }
    if (!!item.langthContent) {
      for (let i = 0, len = item.langthContent?.length; i < len; i++) {
        let itemRule = item.langthContent[i];
        if (!itemRule.imgs?.length) {
          itemRule.imgs = [];
        }
        for (let j = 0, len = itemRule.imgs.length; j < len; j++) {
          let _it = itemRule.imgs[j];
          let imgUrl = await import(`@assets/${_it.url}`);
          _it.url = imgUrl.default;
          // console.log('imgUrl', imgUrl, imgUrl.default, _it); // /extensionBundleCode/img/content_img_1.48dda63..png
        }
      }
    }
    if (!!item.activeRuleContent) {
      for (let i = 0, len = item.activeRuleContent?.length; i < len; i++) {
        let itemRule = item.activeRuleContent[i];
        if (!itemRule.imgs?.length) {
          itemRule.imgs = [];
        }
        for (let j = 0, len = itemRule.imgs.length; j < len; j++) {
          let _it = itemRule.imgs[j];
          let imgUrl = await import(`@assets/${_it.url}`);
          _it.url = imgUrl.default;
          // console.log('imgUrl', imgUrl, imgUrl.default, _it); // /extensionBundleCode/img/content_img_1.48dda63..png
        }
      }
    }

    // viewData.status.ruleActivity = true;
  }
  return Promise.resolve(configs);
}

// 玩家消息-参与文案 whatsapp message
const whatsppMessage = {
  data: {
    "en": {
      "message"({ code = "" }) {
        return `I'm joining the MLBB GOLDEN MONTH bonus sharing event to win 🎁 amazing rewards including $1,000 cash, OPPO phone, 100,000 MLBB Diamonds, and an exclusive skin!\nUse My Code: ${code}`
      }
    },
    "my": {
      "message"({ code = "" }) {
        return `Saya menyertai acara berkongsi bonus MLBB GOLDEN MONTH untuk memenangi 🎁 ganjaran hebat termasuk wang tunai $1,000, telefon OPPO, 100,000 Berlian MLBB, dan juga skin eksklusif!\nGuna Kod Saya: ${code}`
      }
    },
    "id": {
      "message"({ code = "" }) {
        return `Aku ikut event berbagi bonus MLBB GOLDEN MONTH supaya bisa menang 🎁 hadiah keren termasuk uang tunai $1.000, HP OPPO, 100.000 Diamond MLBB, dan skin eksklusif!\nGunakan Kode Punyaku: ${code}`
      }
    }
  }
}
export function queryWhatsppMessageLang(langType = "02") {
  return whatsppMessage.data[LANGUAGE_MODE[langType]];
}

// 活动规则
const invitationActivityRules = {
  data: {
    "en": {
      "activeRuleContentTitleImg": "en/rule-tit.png",
      "activeRuleContent": [{
        "text"() {
          return `This event is organized by Moonton. Please read the event rules and related terms carefully before participating. By participating in this event, you acknowledge that you have read, understood, and agreed to all content in the event rules and terms.<br />· 1 Min 1 Phone for Free In-Game：From 2/28 to 4/1, participate in the in-game Dazzling Golden Spin event for a chance to win an OPPO phone. During specified periods, one OPPO phone will be given away every minute, for a total of 10,000 phones. (Please check the in-game event rules for details.)<br />· WhatsApp Event - Phone Prize Pool Expanded 100X: After 03/21, the OPPO phone prize pool will increase from 16 to 1600 phones, greatly boosting the winning probability. Players who met the drawing requirements before 03/21 are also included in the draw. (Winners will be randomly selected before 04/08/2025, and notified via WhatsApp. If you don't receive a winning notification, it means you haven't won. The list of winners will be published on the event page before 04/09.)`
        }
      }, {
        "text": "This event is currently only available in Malaysia, Indonesia, Philippines, Singapore. Only players with WhatsApp account prefixes (+60, +62, +63, +65) are eligible to participate."
      }, {
        "text": "Event Period: 03/01/2025 00:00:00 - 03/31/2025 23:59:59 (UTC-8). Rewards are limited. The event may end early if all rewards have been claimed."
      }, {
        "text": `Players who participate in the team-up event via the MLBB WhatsApp business account and successfully invite the specified number of friends will receive corresponding rewards.<br/>[Invite 3 Friends]: Guaranteed Premium Chest containing one of the following items: Tigreal "Lightborn - Defender" ×1, Chang'e "Vine Cradle" ×1, Nana "Mecha Baby" ×1, Grock "Codename: Rhino" ×1, Gusion "Soul Revelation" ×1, Double Exp Card ×1, Epic Skin Trial Pack (1-Day) ×1, Skin Trial Pack (1-Day) ×1, Ticket ×5, Magic Dust ×5, Magic Dust ×1, and a chance to win $100 Cash.<br/>[Invite 6 Friends]: Guaranteed a Stellar Chest containing one of the following items: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Lunox "Eyes of Eternity" ×1, Kimmy "Frost Wing" ×1, Masha "Dragon Armor" ×1, Double Exp Card ×1, Epic Skin Trial Pack (1-Day) ×1, Skin Trial Pack (1-Day) ×1, Ticket ×5, Magic Dust ×5, Magic Dust ×1, and a chance to win an OPPO Phone.(Quantity: 1600)<br/>[Invite 9 Friends]: Guaranteed a Moonlight Chest containing one of the following items: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Fanny "Lightborn - Ranger" ×1, Harley "Great Inventor" ×1, Grock "Codename: Rhino" ×1, Gusion "Soul Revelation" ×1, Double Exp Card ×1, Epic Skin Trial Pack (1-Day) ×1, Skin Trial Pack (1-Day) ×1, Ticket ×5, Magic Dust ×5, Magic Dust ×1,  and a chance to upgrade the reward to a 100,000 in-game Diamonds pack.<br/>[Invite 12 Friends]: Guaranteed an Apex Chest containing one of the following items: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Fanny "Lightborn - Ranger" ×1, Harith "Lightborn - Inspirer" ×1, Irithel "Hellfire" ×1, Grock "Codename: Rhino" ×1, Gusion "Soul Revelation" ×1, Double Exp Card ×1, Epic Skin Trial Pack (1-Day) ×1, Skin Trial Pack (1-Day) ×1, Ticket ×5, Magic Dust ×5, Magic Dust ×1,and a chance to win $500 Cash.<br/>[Invite 15 Friends]: Guaranteed a Glory Chest containing one of the following items: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Fanny "Lightborn - Ranger" ×1, Harith "Lightborn - Inspirer" ×1, Granger "Lightborn - Overrider" ×1, Harley "Great Inventor" ×1, Irithel "Hellfire" ×1, Double Exp Card ×1, Epic Skin Trial Pack (1-Day) ×1, Skin Trial Pack (1-Day) ×1, Ticket ×5, Magic Dust ×5, Magic Dust ×1, and a chance to win $1,000 Cash.<br/><br/>In-game rewards will be distributed via CDK. Please watch for WhatsApp message notifications. Copy the CDK and redeem it in-game by tapping your profile picture in the top-left corner, and then accessing Settings in the top-right.Each CDK can only be redeemed once and cannot be used multiple times. Resale and scalping are strictly prohibited. The organizer reserves the right to invalidate any CDK suspected of being resold or obtained illegally and may take legal action if necessary.<br/><br/>For cash and physical prizes, winners will be randomly selected from eligible players on 04/08/2025 and contacted via WhatsApp. If you receive a winning notification, please submit the required information within 15 days, or the prize will be forfeited. If you don't receive a notification, it means you were not selected as a winner. The list of winners will be announced on this page on 04/09. Event rewards will be distributed in accordance with local laws and regulations. If physical/cash rewards cannot be distributed in certain regions, rewards will be converted to Diamonds based on their value and sent to your in-game account.`
      }, {
        "text": "Each player has only one opportunity to help a friend, and the invitation is only considered successful after tapping the friend's link and sending a help message to MLBB."
      }, {
        "text": "The organizer reserves the right to make supplementary interpretations of the event rules to the maximum extent permitted by law. If you have any questions, please contact Customer Service using the button at the top of the in-game main interface."
      }],
      "activeRuleWinningInfoContent": {
        "columns": [{
          thTitle: "WhatsApp Account"
        }, {
          thTitle: "WhatsApp Name"
        }, {
          thTitle: "Prize"
        }],
        "dataSource": []
      }
    },
    "my": {
      "activeRuleContentTitleImg": "my/rule-tit.png",
      "activeRuleContent": [{
        "text": "Acara ini dianjurkan oleh Moonton. Sila baca peraturan acara dan terma berkaitan dengan teliti sebelum menyertai. Dengan menyertai acara ini, anda mengakui bahawa anda telah membaca, memahami dan bersetuju dengan semua kandungan dalam peraturan dan terma acara. <br/>·1 Minit 1 Telefon Percuma dalam Permainan：Dari 28/2 hingga 1/4, sertai acara Dazzling Golden Spin dalam permainan untuk berpeluang memenangi telefon OPPO. Semasa tempoh tertentu, satu telefon OPPO akan diberikan setiap minit, dengan jumlah keseluruhan 10,000 buah telefon. (Sila semak peraturan acara dalam permainan untuk butiran lanjut.)<br />· Acara WhatsApp - Hadiah Terkumpul Telefon Diperluaskan 100X: Selepas 21/03, hadiah terkumpul telefon OPPO akan meningkat daripada 16 kepada 1,600 telefon, sangat meningkatkan kebarangkalian kemenangan. Pemain yang telah memenuhi syarat cabutan sebelum 21/03 juga akan disertakan dalam cabutan. (Pemenang akan dipilih secara rawak sebelum 08/04/2025, dan dimaklumkan melalui WhatsApp. Jika anda tidak menerima pemakluman kemenangan, ini bermakna anda tidak menang. Senarai pemenang akan diumumkan di halaman acara sebelum 09/04.)"
      }, {
        "text": "Acara ini hanya tersedia di Malaysia, Indonesia, Filipina dan Singapura buat masa ini. Hanya pemain dengan awalan nombor akaun WhatsApp (+60, +62, +63, +65) layak untuk menyertai."
      }, {
        "text": "Tempoh Acara: 01/03/2025 00:00:00 - 31/03/2025 23:59:59 (UTC-8). Ganjaran adalah terhad. Acara mungkin tamat lebih awal sekiranya semua ganjaran telah dituntut."
      }, {
        "text": `Pemain yang menyertai acara berpasukan melalui akaun perniagaan WhatsApp MLBB dan berjaya menjemput sejumlah rakan tertentu akan menerima ganjaran yang sepadannya.<br/>[Jemput 3 Rakan]: Dijamin Peti Premium yang mengandungi satu daripada item berikut: Tigreal "Lightborn - Defender" ×1, Chang'e "Vine Cradle" ×1, Nana "Mecha Baby" ×1, Grock "Codename: Rhino" ×1, Gusion "Soul Revelation" ×1, Kad EXP Berganda ×1, Pek Percubaan Skin Epic (1 Hari) ×1, Pek Percubaan Skin (1 Hari) ×1, Tiket ×5, Magic Dust ×5, Magic Dust ×1, dan peluang untuk memenangi wang tunai $100.<br/>[Jemput 6 Rakan]: Dijamin Peti Stellar yang mengandungi satu daripada item berikut: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Lunox "Eyes of Eternity" ×1, Kimmy "Frost Wing" ×1, Masha "Dragon Armor" ×1, Kad EXP Berganda ×1, Pek Percubaan Skin Epic (1 Hari) ×1, Pek Percubaan Skin (1 Hari) ×1, Tiket ×5, Magic Dust ×5, Magic Dust ×1, dan peluang untuk memenangi Telefon OPPO.(Kuantiti: 1600)<br/>[Jemput 9 Rakan]: Dijamin Peti Moonlight yang mengandungi satu daripada item berikut: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Fanny "Lightborn - Ranger" ×1, Harley "Great Inventor" ×1, Grock "Codename: Rhino" ×1, Gusion "Soul Revelation" ×1, Kad EXP Berganda ×1, Pek Percubaan Skin Epic (1 Hari) ×1, Pek Percubaan Skin (1 Hari) ×1, Tiket ×5, Magic Dust ×5, Magic Dust ×1, dan peluang untuk menaik taraf ganjaran ke 100,000 pek Berlian dalam permainan.<br/>[Jemput 12 Rakan]: Dijamin Peti Apex yang mengandungi satu daripada item berikut: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Fanny "Lightborn - Ranger" ×1, Harith "Lightborn - Inspirer" ×1, Irithel "Hellfire" ×1, Grock "Codename: Rhino" ×1, Gusion "Soul Revelation" ×1, Kad EXP Berganda ×1, Pek Percubaan Skin Epic (1 Hari) ×1, Pek Percubaan Skin (1 Hari) ×1, Tiket ×5, Magic Dust ×5, Magic Dust ×1, dan peluang untuk memenangi wang tunai $500.<br/>[Jemput 15 Rakan]: Dijamin Peti Glory yang mengandungi satu daripada item berikut: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Fanny "Lightborn - Ranger" ×1, Harith "Lightborn - Inspirer" ×1, Granger "Lightborn - Overrider" ×1, Harley "Great Inventor" ×1, Irithel "Hellfire" ×1, Kad EXP Berganda ×1, Pek Percubaan Skin Epic (1 Hari) ×1, Pek Percubaan Skin (1 Hari) ×1, Tiket ×5, Magic Dust ×5, Magic Dust ×1, dan peluang untuk memenangi wang tunai $1,000.<br/><br/>Ganjaran dalam permainan akan diedarkan melalui CDK. Sila lihat mesej notifikasi WhatsApp anda. Salin CDK dan tebus dalam permainan dengan menekan gambar profil anda pada sudut kira atas, dan kemudian akses Tetapan pada sudut kanan atas. Setiap CDK hanya bisa ditukarkan sekali dan tidak bisa digunakan berulang kali. Dilarang menjual kembali CDK. Penyelenggara berhak untuk membatalkan CDK yang dicurigai dijual kembali atau didapatkan secara ilegal, serta berhak untuk mengambil tindakan hukum jika diperlukan.<br/><br/>Untuk wang tunai dan hadiah fizikal, pemenang akan dipilih secara rawak daripada pemain yang layak pada 08/04/2025 dan akan dihubungi melalui WhatsApp. Jika anda menerima notifikasi kemenangan, sila hantar maklumat yang diperlukan dalam masa 15 hari, atau hadiah anda akan dibatalkan. Jika anda tidak menerima notifikasi, ia bermakna anda tidak terpilih sebagai pemenang. Senarai pemenang akan diumumkan di halaman ini pada 09/04. Ganjaran acara akan diagihkan mengikut undang-undang dan peraturan tempatan. Jika ganjaran fizikal/wang tunai tidak dapat diagihkan di sesetengah rantau, ganjaran akan ditukar kepada Berlian berdasarkan nilainya dan dihantar ke akaun dalam permainan anda.`
      }, {
        "text": "Setiap pemain hanya mempunyai satu peluang untuk membantu rakan, dan jemputan hanya dianggap berjaya selepas menekan pautan rakan dan menghantar mesej bantuan ke MLBB. "
      }, {
        "text": "Penganjur berhak untuk meminda peraturan acara sehingga yang dibenarkan oleh undang-undang. Jika anda mempunyai sebarang pertanyaan, sila hubungi Khidmat Pelanggan dengan menggunakan butang di atas antara muka utama dalam permainan."
      }],
      "activeRuleWinningInfoContent": {
        "columns": [{
          thTitle: "Akaun WhatsApp"
        }, {
          thTitle: "Nama WhatsApp"
        }, {
          thTitle: "Hadiah"
        }],
        "dataSource": []
      }
    },
    "id": {
      "activeRuleContentTitleImg": "id/rule-tit.png",
      "activeRuleContent": [{
        "text"() {
          return `Event ini diselenggarakan oleh Moonton. Harap baca peraturan dan ketentuan dengan saksama sebelum berpartisipasi. Dengan berpartisipasi di event ini, kamu mengakui bahwa kamu sudah membaca, memahami, dan menyetujui seluruh isi peraturan dan ketentuan event.<br/>·1 Menit 1 HP Gratis Dalam Game：Dari 28/02 sampai 01/04, ikuti event dalam game Dazzling Golden Spin untuk kesempatan memenangkan HP OPPO. Selama periode yang ditentukan, satu HP OPPO akan diberikan setiap menit, dengan total 10.000 HP. (Cek detail aturan event dalam game.)<br />· Event WhatsApp - Hadiah HP Diperbanyak 100X: Setelah 21/03, hadiah HP OPPO akan meningkat dari 16 menjadi 1.600 HP, peluang menang meningkat drastis! Player yang memenuhi syarat draw sebelum 21/03 juga akan dimasukkan ke dalam undian. (Pemenang akan dipilih secara acak sebelum 08/04/2025, dan akan diberi tahu melalui WhatsApp. Jika kamu tidak menerima notifikasi menang, berarti kamu belum menang. Daftar pemenang akan dipublikasikan di halaman event sebelum 09/04).`
        }
      }, {
        "text": "Event ini hanya tersedia di Malaysia, Indonesia, Filipina, dan Singapura. Hanya player dengan awalan akun WhatsApp (+60, +62, +63, +65) yang memenuhi syarat untuk berpartisipasi."
      }, {
        "text": "Periode Event: 01/03/2025 00:00:00 - 31/03/2025 23:59:59 (UTC-8). Hadiahnya terbatas. Event mungkin akan berakhir lebih awal jika semua hadiah sudah diklaim."
      }, {
        "text": `Player yang berpartisipasi di event main bersama melalui akun bisnis WhatsApp MLBB dan berhasil mengundang teman dengan nomor yang ditentukan akan menerima hadiah yang sesuai.<br/>[Undang 3 Teman]: Dijamin mendapatkan Premium Chest yang berisi salah satu item berikut: Tigreal "Lightborn - Defender" ×1, Chang'e "Vine Cradle" ×1, Nana "Mecha Baby" ×1, Grock "Codename: Rhino" ×1, Gusion "Soul Revelation" ×1, Double EXP Card ×1, Epic Skin Trial Pack (1 Hari) ×1, Skin Trial Pack (1 Hari) ×1, Ticket ×5, Magic Dust ×5, Magic Dust ×1, dan kesempatan memenangkan Uang Tunai $100.<br/>[Undang 6 Teman]: Dijamin mendapatkan Stellar Chest yang berisi salah satu item berikut: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Lunox "Eyes of Eternity" ×1, Kimmy "Frost Wing" ×1, Masha "Dragon Armor" ×1, Double EXP Card ×1, Epic Skin Trial Pack (1 Hari) ×1, Skin Trial Pack (1 Hari) ×1, Ticket ×5, Magic Dust ×5, Magic Dust ×1, dan kesempatan memenangkan HP OPPO.(Jumlah: 1600)<br/>[Undang 9 Teman]: Dijamin mendapatkan Moonlight Chest yang berisi salah satu item berikut: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Fanny "Lightborn - Ranger" ×1, Harley "Great Inventor" ×1, Grock "Codename: Rhino" ×1, Gusion "Soul Revelation" ×1, Double EXP Card ×1, Epic Skin Trial Pack (1 Hari) ×1, Skin Trial Pack (1 Hari) ×1, Ticket ×5, Magic Dust ×5, Magic Dust ×1, dan kesempatan meningkatkan hadiah menjadi pack 100.000 Diamond dalam game.<br/>[Undang 12 Teman]: Dijamin mendapatkan Apex Chest yang berisi salah satu item berikut: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Fanny "Lightborn - Ranger" ×1, Harith "Lightborn - Inspirer" ×1, Irithel "Hellfire" ×1, Grock "Codename: Rhino" ×1, Gusion "Soul Revelation" ×1, Double EXP Card ×1, Epic Skin Trial Pack (1 Hari) ×1, Skin Trial Pack (1 Hari) ×1, Ticket ×5, Magic Dust ×5, Magic Dust ×1, dan kesempatan memenangkan Uang Tunai $500.<br/>[Undang 15 Teman]: Dijamin mendapatkan Glory Chest yang berisi salah satu item: Tigreal "Lightborn - Defender" ×1, Alucard "Lightborn - Striker" ×1, Fanny "Lightborn - Ranger" ×1, Harith "Lightborn - Inspirer" ×1, Granger "Lightborn - Overrider" ×1, Harley "Great Inventor" ×1, Irithel "Hellfire" ×1, Double EXP Card ×1, Epic Skin Trial Pack (1 Hari) ×1, Skin Trial Pack (1 Hari) ×1, Ticket ×5, Magic Dust ×5, Magic Dust ×1, dan kesempatan memenangkan Uang Tunai $1.000.<br/><br/>Hadiah dalam game akan dikirim melalui CDK. Harap perhatikan notifikasi pesan WhatsApp. Salin CDK dan tukarkan dalam game dengan mengetuk gambar profilmu di pojok kiri atas, lalu akses Pengaturan di kanan atas. <br/>Setiap CDK hanya boleh ditebus sekali dan tidak boleh digunakan berulang kali. Penjualan semula dan pembelian untuk dijual semula adalah dilarang sama sekali. Pihak penganjur berhak untuk membatalkan mana-mana CDK yang disyaki dijual semula atau diperoleh secara haram, dan boleh mengambil tindakan undang-undang jika perlu.<br/><br/>Untuk uang tunai dan hadiah fisik, pemenang akan dipilih secara acak dari player yang memenuhi syarat pada 08/04/2025 dan akan dihubungi melalui WhatsApp. Jika kamu menerima notifikasi menang, harap isi informasi yang diperlukan dalam 15 hari, atau hadiah akan hangus. Jika kamu tidak menerima notifikasi, berarti kamu tidak terpilih sebagai pemenang. Daftar pemenang akan diumumkan di halaman ini pada 09/04.<br/>Hadiah event akan dikirimkan sesuai dengan hukum dan peraturan umum yang berlaku. Jika hadiah fisik/uang tunai tidak bisa dikirimkan pada wilayah tertentu, hadiah akan dikonversi menjadi Diamond berdasarkan nilainya dan dikirim ke akun dalam game.`
      }, {
        "text": "Setiap player hanya punya satu kesempatan untuk membantu teman, dan undangan akan dianggap berhasil setelah mengetuk link dari teman dan mengirim pesan bantuan ke MLBB."
      }, {
        "text": "Penyelenggara berhak memberikan penjelasan tambahan terhadap peraturan event sejauh diizinkan oleh hukum yang berlaku. Jika kamu punya pertanyaan, harap hubungi Customer Service menggunakan tombol atas di interface utama dalam game."
      }],
      "activeRuleWinningInfoContent": {
        "columns": [{
          thTitle: "Akun WhatsApp"
        }, {
          thTitle: "Nama WhatsApp"
        }, {
          thTitle: "Hadiah"
        }],
        "dataSource": []
      }
    }
  }
}
export function queryActivityRules(langType = "01") {
  // console.log('LANGUAGE_MODE[langType]', LANGUAGE_MODE[langType]);
  return invitationActivityRules.data[LANGUAGE_MODE[langType]];
}


// 切换语言弹窗
export const switchLangthModal = {
  data: {
    "en": {
      "langthTitleImg": "",
      "langthContent": [{
        "text": "The current page language has been changed to [English]."
      }, {
        "text": "Do you want to switch the active message language of MLBB WhatsApp at the same time?"
      }]
    },
    "my": {
      "langthTitleImg": "",
      "langthContent": [{

        "text": "Bahasa laman telah ditukar kepada [Bahasa Melayu]."
      }, {
        "text": "Adakah anda ingin menukar bahasa mesej WhatsApp MLBB juga?"
      }]
    },
    "id": {
      "langthTitleImg": "",
      "langthContent": [{
        "text": "Bahasa halaman diganti ke [Bahasa Indonesia]."
      }, {
        "text": "Apakah kamu ingin ganti bahasa pesan WhatsApp MLBB juga?"
      }]
    }
  }
}
export function querySwitchLangthModal(langType = "02") {
  return switchLangthModal.data[LANGUAGE_MODE[langType]];
}