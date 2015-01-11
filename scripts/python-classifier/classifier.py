

'''
Before running this, you will need to create an IMAGE_FILE in the current directory
'''

import numpy as np
caffe_python = '/opt/caffe/python'
import sys
sys.path.insert(0, caffe_python)

import caffe
import glob 
import json
import os

MODEL_FILE = 'classifier.prototxt'
PRETRAINED = 'caffe.model'

# TODO: this needs to be passed in!  It can be calculated from prototxt file 
# by taking the inverse of the scale parameter.
# See https://github.com/tleyden/caffe/blob/047d3dac8b25b0edf452e53aefd33cb47d8042d3/examples/alphabet_classification/alpha_train_text.prototxt#L13
RAW_SCALE = 255  

# TODO: calculate image_dims
image_dims=(28, 28)

# TODO: this should be parameterized
COLOR=False

net = caffe.Classifier(MODEL_FILE, 
                       PRETRAINED,
                       raw_scale=RAW_SCALE,
                       image_dims=image_dims)

net.set_phase_test()
net.set_mode_cpu()

# go into images subdir
os.chdir("images")

# loop over all files in directory that are named image*
image_filenames = glob.glob('*')
images = []

for image_filename in image_filenames:

    if not os.path.exists(image_filename):
        break

    # load each image and add to images array
    images.append(caffe.io.load_image(image_filename, color=COLOR))

if len(images) == 0:
    raise Exception("no images")

# go up a directory so we don't write result in the images dir
os.chdir("..")

predictions = net.predict(images)

result = {}

for image_index in xrange(len(image_filenames)):

    image_filename = image_filenames[image_index]
    prediction = predictions[image_index]
    result[image_filename] = prediction.argmax()
    # print 'prediction shape:', prediction.shape

# write json result
f = open('result.json', 'w')
json.dump(result, f)

print "Current dir:"
os.system("pwd")
print "Output saved to result.json"

