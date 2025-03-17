(defcolumns (ARR :i16 :array [2]))
(definterleaved Z ([ARR 1] [ARR 2]))
(defconstraint c1 () (== 0 Z))
