;;error:4:23-24:conflicting context
(defcolumns (X :i16) (Y :i16))
(definterleaved A (X Y))
(defproperty p1 (== A X))
