;;error:13:17-18:array index out-of-bounds
;;
(defcolumns
    (BIT :binary@prove :array [3])
    (ARG :i16))

(defconstraint bits ()
  (== ARG
     (+
      (* 1 [BIT 1])
      (* 2 [BIT 2])
      (* 4 [BIT 3])
      (* 8 [BIT 4]))))
