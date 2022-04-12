ROOM_ID=274f2e2f-5675-45ab-a32f-a74cf07e3ad6

. "$(pwd)"/scripts/get-secrets.sh prd && \

bin/livekit-load-tester \
    --url wss://leap.livekit.cloud \
    --api-key $LIVE_KIT_API_KEY \
    --api-secret $LIVE_KIT_API_SECRET \
    --room "$ROOM_ID" \
    --identity-range 13173-13183 \
    # --publishers 5
