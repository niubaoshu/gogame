import tensorflow as tf
import tensorflow.keras as keras


def read_file_x(path, n):
    """

    :param path:
    :param n:
    :return:
    """
    with open(path, 'rb') as f:
        d = f.read(361 * n)
        return list(d)


def read_file_y(path, n):
    """

    :param path:
    :param n:
    :return:
    """
    with open(path, 'rb') as f:
        d = f.read(2 * n)
        return list(d)


def get_data(path, n):
    """

    :param path:
    :param n:
    """
    x = read_file_x(path + ".datax", n)
    y = read_file_y(path + ".datay", n)

    x = tf.reshape(x, (n, 19, 19))-1
    y = tf.cast(tf.tensordot(tf.reshape(y, (n,2)), (256, 1), axes=1)-361, dtype=float)
    # y = tf.one_hot(tf.tensordot(tf.reshape(y, (n,2)), (256, 1), axes=1),722)
    # y = tf.tensordot(tf.reshape(y, (n,2)), (256, 1), axes=1)

    return x, y


model = keras.models.Sequential()
model.add(
    keras.layers.Conv2D(64, kernel_size=3, activation=tf.nn.relu, input_shape=(19, 19, 1), strides=1, padding="same"))
model.add(keras.layers.MaxPool2D((2, 2)))
model.add(keras.layers.Conv2D(64, kernel_size=5, activation=tf.nn.relu, strides=1, padding="same"))
model.add(keras.layers.MaxPool2D((2, 2)))
# model.add(keras.layers.Conv2D(32,(7,7),activation=tf.nn.relu,strides=1,padding="same"))
# model.add(keras.layers.MaxPool2D((2, 2)))
# model.add(keras.layers.MaxPool2D((2,2)))
# model.add(keras.layers.Conv2D(128,(3,3),activation=tf.nn.relu,input_shape=(19,19,1),strides=1,padding="same"))
# model.add(keras.layers.MaxPool2D((2,2)))
# model.add(keras.layers.Flatten())
# model.add(keras.layers.Dense(32,activation=keras.activations.relu))
# model.add(keras.layers.Dense(10,activation=keras.activations.relu))
# model.add(keras.layers.BatchNormalization())
model.add(keras.layers.Flatten())
# model.add(keras.layers.Dense(1000,activation=keras.activations.relu))
# tf.keras.layers.Dropout(0.2),
model.add(keras.layers.Dense(1))

model.compile(optimizer=keras.optimizers.Adam(), loss=keras.losses.MeanSquaredError())
model.build((None,19,19))
model.summary()

n = 100000
batch_size = 200

x, y = get_data("../main/goGame_train", n)

# model.load_weights("weights")
print(x.shape,tf.nn.moments(y, 0))

model.fit(x, y, epochs=20, batch_size=batch_size)
x, y = get_data("../main/goGame_test", n)
print(x.shape,tf.nn.moments(y, 0))
model.evaluate(x, y,batch_size=batch_size, verbose=2)

# model.save_weights("weights")

# print(x.shape, y,x[:1][:][:].shape)
# result = model.predict(x)
# for i in range(0,100):
#     print(result[i],y[i])