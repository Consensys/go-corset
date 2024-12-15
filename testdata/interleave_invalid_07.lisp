;;error:2:1-2:blah
(defcolumns X Y)
(definterleaved A (X Y))
(defproperty p1 (+ A X))
