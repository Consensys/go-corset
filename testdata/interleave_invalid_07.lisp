;;error:4:22-23:conflicting context
(defcolumns X Y)
(definterleaved A (X Y))
(defproperty p1 (+ A X))
