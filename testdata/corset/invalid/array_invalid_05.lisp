;;error:10:17-18:array index out-of-bounds
;;
(defcolumns
    (BIT :binary@prove :array [4])
    (ARG :i16))

(defconstraint bits ()
  (== ARG
     (+
      (* 1 [BIT 0])
      (* 2 [BIT 1])
      (* 4 [BIT 2])
      (* 8 [BIT 3]))))
