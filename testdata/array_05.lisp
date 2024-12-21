(defcolumns
    (BIT :binary@prove :array [0:3])
    (ARG :i16@loob))

(defconstraint bits ()
  (- ARG
     (+
      (* 1 [BIT 0])
      (* 2 [BIT 1])
      (* 4 [BIT 2])
      (* 8 [BIT 3]))))
