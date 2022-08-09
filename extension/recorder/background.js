// @ts-nocheck
/* global chrome, MediaRecorder, FileReader */

const recorders = {};

/**
 * 
 * @param {*} blob 
 * @returns 
 */
function blobToBase64(blob) {
	return new Promise((resolve, _) => {
		const reader = new FileReader();
		reader.onloadend = () => resolve(reader.result);
		reader.readAsDataURL(blob);
	});
}

function START_RECORDING(params) {
	const { index, video, audio, frameSize, audioBitsPerSecond, videoBitsPerSecond, bitsPerSecond, mimeType, videoConstraints } = params
	console.log("START_RECORDING_PARAMS:", params)

	chrome.tabCapture.capture(
		{
			audio,
			video,
			audioConstraints: {
				mandatory: {
					chromeMediaSource: 'tab',
					echoCancellation: true
				}
			},
			videoConstraints: {
				mandatory: {
					chromeMediaSource: 'tab',
				}
			}
		},
		(stream) => {
			if (!stream) {
				console.warn("No stream found!")
			};

			recorder = new MediaRecorder(stream, {
				ignoreMutedMedia: true,
				videoMaximizeFrameRate: true,
				audioBitsPerSecond,
				videoBitsPerSecond,
				bitsPerSecond,
				mimeType,
			});

			recorders[index] = recorder;

			recorder.onerror = (event) => {
				console.error(`error recording stream: ${event.error.name}`)
				console.error(event)
			};

			recorder.ondataavailable = async function (event) {
				if (event.data.size > 0) {
					const b = new Blob([event.data])
					const base64Str = await blobToBase64(b)

					if (window.sendWholeData) {
						window.sendWholeData({
							type: index,
							chunk: base64Str,
						});
					}
				}
			};

			recorder.onerror = () => {
				recorder.stop();
			}

			recorder.onstop = () => {
				try {
					const tracks = stream.getTracks();
					tracks.forEach(function (track) {
						track.stop();
					});
				} catch (error) { }
			};

			stream.oninactive = () => {
				try {
					recorder.stop();
				} catch (error) { }
			};

			// start recording
			console.log("started recording with frameSize:", frameSize)
			recorder.start(frameSize);
		}
	);
}

function STOP_RECORDING(index) {
	if (!recorders[index]) return;
	recorders[index].stop();
}
