from os import getcwd, listdir, walk
from sys import argv, exit
from typing import List

import moviepy.editor as mp  # type: ignore

from cli import Cli, Params


def main(params: Params):

    videos = []
    pwd = getcwd()

    for (dirpath, dirnames, filenames) in walk(f"{pwd}/video"):
        for file_name in filenames:
            if (file_name != '.gitkeep'):
                v = mp.VideoFileClip(f"{pwd}/video/{file_name}")
            videos.append(v)

    index = params.get('index')
    prefix = params.get('prefix')
    height = params.get('height')
    kbps = params.get('kbps')
    fps = params.get('fps')
    out = params.get('out')

    for video in videos:
        video_name = f"{out}/{prefix}{index}" 
        
        video_name = f"{video_name}_{height}_{kbps}_{fps}.h264"
        
        current_width, current_height = video.size
        ratio = current_width / current_height

        new_width = round(height * ratio)

        video.set_fps(video.fps)
        video.resize((new_width, height))
        video.loop(duration=video.duration)
        video.write_videofile(filename=video_name, fps=fps, codec="libx264", audio=False)

        index += 1


if __name__ == "__main__":
    args = argv[1:]
    app = Cli()

    if len(args) == 0:
        app.help()

    for arg in args:
        split = arg.split("=")
        
        if len(split) != 2:
            app.help()
            exit(1)
        
        key, value = split
        app.parse(key, value)

    p = app.get_params()

    main(p)

