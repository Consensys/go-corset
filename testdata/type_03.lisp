(defcolumns (X16 :u16)
  (D8 :u8))

(defconstraint sorted () (- D8 (- X16 (shift X16 -1))))
