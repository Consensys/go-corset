(defcolumns (X16 :i16@prove)
  (D8 :i8@prove))

(defconstraint sorted () (- D8 (- X16 (shift X16 -1))))
