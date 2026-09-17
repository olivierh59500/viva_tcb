# VIVA TCB

Remake Go/Ebitengine de l’écran VIVA TCB. Le même cœur de jeu alimente la
version ordinateur et l’application Android.

La musique `assets/music.ym` est synthétisée par `ym-player` à 48 kHz, puis
envoyée à l’audio Ebitengine sous forme PCM 16 bits stéréo. L’initialisation de
l’audio est différée au premier `Update` pour respecter le cycle de vie Android.

## Ordinateur

```sh
go run ./cmd/vivatcb
```

## Android

Pré-requis validés : Go 1.25 ou ultérieur, JDK 17, Android SDK/API 36, NDK r28
et un appareil Android autorisé via USB.

Pour construire l’AAR arm64 et l’APK debug, installer l’application puis la
lancer sur l’unique appareil connecté :

```sh
./scripts/run-android.sh
```

Artefacts générés :

```text
android/app/libs/vivatcb.aar
android/app/build/outputs/apk/debug/app-debug.apk
```

L’application utilise l’identifiant `com.olivierh.vivatcb`, le mode immersif
et les deux orientations paysage. Ebitengine conserve le canvas logique
768 × 540 et le centre sans déformation sur les écrans plus larges.

## Vérifications

```sh
go test ./...
go test -race ./...
go vet ./...
golangci-lint run
```
