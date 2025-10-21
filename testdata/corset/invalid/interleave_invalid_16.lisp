;;error:3:17-25:incompatible type (u32)
(defcolumns (X :i32) (Y :i16))
(definterleaved (Z :i16) (X Y))
