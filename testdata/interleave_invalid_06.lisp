;;error:2:1-2:blah
(defcolumns X Y)
(definterleaved A (X Y))
(defconstraint c1 () (+ A X))
