(function (pageId, contraints) {
    contraints = contraints || {
        audio: true,
        video: true,
        mimeType: "video/webm;codecs=vp8,opus",
        audioBitsPerSecond: 128000,
        videoBitsPerSecond: 2500000,
        bitsPerSecond: 8000000,
        frameSize: 1000,
    }
    // Index for recording
    contraints.index = pageId
    window.START_RECORDING(contraints)
})
