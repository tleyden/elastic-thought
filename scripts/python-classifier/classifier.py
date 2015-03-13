

"""
Before running this, you will need to create an IMAGE_FILE in the current directory
"""

import numpy as np
caffe_python = '/opt/caffe/python'
import sys
sys.path.insert(0, caffe_python)

import caffe
import glob 
import json
import os
import getopt
import sys

MODEL_FILE = 'classifier.prototxt'
PRETRAINED = 'caffe.model'

use_gpu = False 
color = False

# This could be calculated from prototxt file by taking the inverse of the scale parameter.
# See https://github.com/tleyden/caffe/blob/047d3dac8b25b0edf452e53aefd33cb47d8042d3/examples/alphabet_classification/alpha_train_text.prototxt#L13
raw_scale = -1
image_width = -1
image_height = -1

options, remainder = getopt.getopt(sys.argv[1:], 's:w:h:cg', ['scale=', 
                                                              'image-width=',
                                                              'image-height=',
                                                              'color',
                                                              'gpu'])

for opt, arg in options:
    if opt in ('-s', '--scale'):
        raw_scale = arg
    elif opt in ('-w', '--image-width'):
        image_width = arg
    elif opt in ('-h', '--image-height'):
        image_height = arg
    elif opt in ('-c', '--color'):
        color = True 
    elif opt in ('-g', '--gpu'):
        use_gpu = True  


if raw_scale == -1 or image_width == -1 or image_height == -1:
    raise Exception("Missing required parameters")

image_dims=(int(image_width), int(image_height))

if use_gpu:
    caffe.set_mode_gpu()
else:
    caffe.set_mode_cpu()


net = caffe.Classifier(MODEL_FILE, 
                       PRETRAINED,
                       raw_scale=int(raw_scale),
                       image_dims=image_dims)


# go into images subdir
os.chdir("images")

# loop over all files in directory that are named image*
image_filenames = glob.glob('*')
images = []

for image_filename in image_filenames:

    if not os.path.exists(image_filename):
        break

    # load each image and add to images array
    images.append(caffe.io.load_image(image_filename, color=color))

if len(images) == 0:
    raise Exception("no images")

# go up a directory so we don't write result in the images dir
os.chdir("..")

predictions = net.predict(images)

result = {}

for image_index in xrange(len(image_filenames)):

    image_filename = image_filenames[image_index]
    prediction = predictions[image_index]
    result[image_filename] = str(prediction.argmax())
    # print 'prediction shape:', prediction.shape

# write json result
f = open('result.json', 'w')
json.dump(result, f)

# debug
json.dumps(result)

# more debug
print "Current dir:"
os.system("pwd")

print "Output saved to result.json"

