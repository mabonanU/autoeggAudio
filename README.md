# ポケモン剣盾 自動色厳選

## 概要
卵を自動孵化し、色違いを厳選します。<br>
Raspberry Pi Zero WとUSBケーブル1本の構成で実現します。<br>

## 準備するもの
・Raspberry Pi Zero W/WH（でなくても、USB On-The-Go対応デバイスならいける？）<br>
・USB A-MicroB データ転送用ケーブル<br>
・WifiでRaspberry piにssh接続できる環境<br>

## 導入手順
### Raspberry Pi Zero W にRaspberry Pi OSをインストール
下記を参考にインストール。<br>
(2020/11/7現在の最新バージョンで動作確認。 Raspberry Pi Imager:1.4 OS:2020-08-20-raspios-buster-armhf-lite)<br>
参考：[USB 1本とSDカードライタだけでできるUSB OTGを用いたRapsberry Pi Zero WH のセットアップ](https://qiita.com/Liesegang/items/dcdc669f80d1bf721c21)<br>
<br>
OSインストール後、Wifi接続でsshできるようになったら、cmdline.txtからload=dwc2,g_etherを削除する。<br>
```sh
$ sudo vi /boot/cmdline.txt
```
<br>

### Raspberry Piをプロコン&USBオーディオとして認識させる設定
参考：[UAC GadgetでNintendo Switchの音声出力をRaspberry Piに取り込む](https://mzyy94.com/blog/2020/04/17/nintendo-switch-audio-uac-gadget/)<br>
```sh
$ sudo vi /etc/modules
```
を実行し、末尾に<br>
```
dwc2
libcomposite
```
を追記する。<br>
終わったら再起動。<br>
```sh
$ sudo reboot
```

起動したら再接続し、[設定用スクリプト](https://gist.github.com/mzyy94/02bcd9d843c77896803c4cd0c4d9b640/raw/aceb75f0deba5166af749ac9007e31a8434f3061/procon_audio.sh)をダウンロードし、実行権限を変更。<br>
```sh
$ wget https://gist.github.com/mzyy94/02bcd9d843c77896803c4cd0c4d9b640/raw/aceb75f0deba5166af749ac9007e31a8434f3061/procon_audio.sh
$ chmod u+x procon_audio.sh
```
スクリプトを実行し、SwitchのUSBポートに接続する。SwitchからUSBオーディオとして認識されればOK。<br>
```sh
$ sudo ./procon_audio.sh
```

### 自動厳選ツールを動作させるための設定
Raspberry PiにPortAudioをインストール。<br>
```sh
$ sudo apt-get install libportaudio2
```
本体をダウンロードし、実行権限を変更。<br>
```sh
$ wget https://github.com/mabonanU/autoeggAudio/raw/main/autoegg
$ chmod u+x autoegg
$ wget https://raw.githubusercontent.com/mabonanU/autoeggAudio/main/run_autoegg.sh
$ chmod u+x run_autoegg.sh
```

## ゲーム側の準備


## 自動厳選実行
SwitchのUSBポートに接続し実行。10秒ほど待つと自動操作が始まります。<br>
```sh
$ sudo ./run_autoegg.sh
```
初回実行時は、
- 空を飛ぶ
- 走って卵回収＆孵化
- 空を飛ぶ
- 産まれたポケモンを逃がす<br>
の1サイクルが問題なく進むことを確認してください。<br>
