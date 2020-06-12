all: build install open log

build:
	@bash ./gradlew build

install:
	@bash ./gradlew installDebug

open:
	@adb shell am start -n dev.fragoso.thdwb/android.app.NativeActivity

clean: 
	@bash ./gradlew cleanToolchain

log:
	@adb logcat --pid=`bash get_pid.sh`
