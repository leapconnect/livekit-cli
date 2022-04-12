# Injects env variables LIVE_KIT_API_KEY and LIVE_KIT_API_SECRET from doppler project environment.
# In order to use this script you need doppler cli to be installed in your system and set up pointint 
# at `leap-api` project

# TODO: save secrets and avoid running doppler if configuration is == to ENV and secrets are already present

ENV=""
LIVE_KIT_API_KEY=""
LIVE_KIT_API_SECRET=""

case "$1" in
    "prd")
    ENV=$1
    ;;
    "stg")
    ENV=$1
    ;;
    "dev")
    ENV=$1
    ;;
    *)
    echo "$1 not a recognize option. options: {prd|stg|dev}"
    exit 1
    ;;
esac

echo "Fetching doppler secrets for env $1..."

export LIVE_KIT_API_SECRET=$(doppler --config ${ENV} secrets get LIVE_KIT_API_SECRET --plain)
export LIVE_KIT_API_KEY=$(doppler --config ${ENV} secrets get LIVE_KIT_API_KEY --plain)
