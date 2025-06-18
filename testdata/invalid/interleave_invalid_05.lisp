;;error:4:22-23:incompatible length multiplier
(defcolumns (X :i16) (Y :i16))
(definterleaved A (X Y))
(definterleaved B (X A))
