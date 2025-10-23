;;error:3:17-25:incompatible type (u32)
(defcolumns (X :i16) (Y :i32))
(definterleaved (Z :i16) (X Y))
