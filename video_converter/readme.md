# Videoconverter

## Quickstart

### Setup python virtual environment

in `video_converter/` folder run 

```
python -m venv .venv
```
You should have Python 3.10 ideally

Activate the environment
```
source .venv/bin/activate
```

Install dependencies

```
pip install -r requirements.txt
```

### Convert videos

Video to be converted in `video/`folder
Default output `dest/` folder

```
python src/main.py
```

To see different options

```
python src/main.py --help
```






Place files that you wanna convert into `video/` folder.