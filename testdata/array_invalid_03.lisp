;;error:4:35-36:expected constant array index
(defcolumns (X :i32) (BIT :i16 :array [4]))

(defconstraint bits () (== 0 [BIT X]))
