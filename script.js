let decodeImg;
let encodeimg;
if (!WebAssembly.instantiateStreaming) {
  // polyfill
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer();
    return await WebAssembly.instantiate(source, importObject);
  };
}

const go = new Go();
let mod, inst;
WebAssembly.instantiateStreaming(fetch("lib.wasm"), go.importObject).then(
  async (result) => {
    mod = result.module;
    inst = result.instance;
    await go.run(inst);
  }
);

async function run() {
  await go.run(inst);
  inst = await WebAssembly.instantiate(mod, go.importObject); // reset instance
}

let p = document.getElementById("code");

function toBase64(a) {
  return btoa(
    a.split().reduce((data, byte) => data + String.fromCharCode(byte), "")
  );
}

async function decodeImage(e) {
  let uDecodePassword = document.getElementById("passwordDecode").value;

  let msg = imageDecode(decodeImg, uDecodePassword);
  document.getElementById("decodeans").innerHTML = msg;
}

async function readURL(image) {
  let reader = new FileReader();
  reader.onload = (e) => {
    let data1 = e.target.result;
    let data = new Uint8Array(data1);

    decodeImg = data;
  };
  reader.readAsArrayBuffer(image.files[0]);
}

async function encodeImage(e) {
  let msg = document.getElementById("msg").value;
  let uPassword = document.getElementById("password").value;
  let yourEncodedImg = imageEncode(decodeImg, msg, uPassword);
  let d = "data:image/png;base64," + yourEncodedImg;
  document.getElementById("finalImage").src = d;
  document.getElementById("download").href = d;
}

const txtarea = document.getElementById("msg");
const txtln = document.getElementById("txtlen");
function updatetxtlen() {
  let data = txtarea.value;
  txtln.innerText = `Length : ${data.length}`;
  // txtln.innerText = `Length : ${data.length}`;
}
